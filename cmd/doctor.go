package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/OrionVesper/tfplan-pretty-output/storage"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose common setup issues",
	Long: `Run a series of health checks against your tfplan-pretty-output setup.

Reports problems with actionable suggestions for each. Useful when something
behaves unexpectedly or when sharing a setup issue with someone else.
`,
	RunE: runDoctor,
}

type checkResult struct {
	Level  string
	Label  string
	Value  string
	Hints  []string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("🩺 Running tfplan-pretty-output health checks...")
	fmt.Println()

	var results []checkResult
	results = append(results,
		checkTerraformBinary(),
		checkTerraformVersion(),
		checkWorkdirWritable(),
		checkStorageRoot(),
		checkProjectIDFile(),
	)
	results = append(results, checkLineages()...)
	results = append(results, checkBrokenLineages()...)

	errs, warns := 0, 0
	for _, r := range results {
		switch r.Level {
		case "ok":
			fmt.Printf("✅ %-30s %s\n", r.Label, r.Value)
		case "info":
			fmt.Printf("ℹ️  %-30s %s\n", r.Label, r.Value)
		case "warn":
			warns++
			fmt.Printf("⚠️  %-30s %s\n", r.Label, r.Value)
			for _, h := range r.Hints {
				fmt.Printf("   → %s\n", h)
			}
		case "err":
			errs++
			fmt.Printf("❌ %-30s %s\n", r.Label, r.Value)
			for _, h := range r.Hints {
				fmt.Printf("   → %s\n", h)
			}
		}
	}

	fmt.Println()
	if errs == 0 && warns == 0 {
		fmt.Println("✨ All checks passed. You're good to go.")
		return nil
	}
	fmt.Printf("✗ %d error(s), %d warning(s). See suggestions above.\n", errs, warns)
	if errs > 0 {
		os.Exit(1)
	}
	return nil
}

func checkTerraformBinary() checkResult {
	path, err := exec.LookPath("terraform")
	if err != nil {
		return checkResult{
			Level: "err", Label: "Terraform binary",
			Value: "not found on PATH",
			Hints: []string{
				"Install: brew install terraform (macOS)",
				"Or download from https://terraform.io",
			},
		}
	}
	return checkResult{Level: "ok", Label: "Terraform binary found:", Value: path}
}

func checkTerraformVersion() checkResult {
	cmd := exec.Command("terraform", "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return checkResult{
			Level: "warn", Label: "Terraform version",
			Value: "could not determine",
			Hints: []string{fmt.Sprintf("Error running 'terraform version': %v", err)},
		}
	}
	first := strings.SplitN(out.String(), "\n", 2)[0]
	first = strings.TrimSpace(first)
	if first == "" {
		first = "(unknown)"
	}
	return checkResult{Level: "ok", Label: "Terraform version:", Value: first}
}

func checkWorkdirWritable() checkResult {
	wd, err := os.Getwd()
	if err != nil {
		return checkResult{Level: "err", Label: "Working directory", Value: "could not determine"}
	}
	probe := filepath.Join(wd, ".tfplan-pretty-output-write-test")
	if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
		return checkResult{
			Level: "err", Label: "Working directory NOT writable:",
			Value: wd,
			Hints: []string{"Check permissions on this directory"},
		}
	}
	_ = os.Remove(probe)
	return checkResult{Level: "ok", Label: "Working directory writable:", Value: wd}
}

func checkStorageRoot() checkResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return checkResult{Level: "err", Label: "Storage root", Value: "could not determine home"}
	}
	root := filepath.Join(home, ".tfplan-pretty-output")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return checkResult{
			Level: "err", Label: "Storage root NOT writable:",
			Value: root,
			Hints: []string{fmt.Sprintf("Try: chmod 755 %s", root)},
		}
	}
	probe := filepath.Join(root, ".write-test")
	if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
		return checkResult{
			Level: "err", Label: "Storage root NOT writable:",
			Value: root,
			Hints: []string{fmt.Sprintf("Try: chmod 755 %s", root)},
		}
	}
	_ = os.Remove(probe)
	return checkResult{Level: "ok", Label: "Storage root writable:", Value: root}
}

func checkProjectIDFile() checkResult {
	wd, err := os.Getwd()
	if err != nil {
		return checkResult{Level: "err", Label: "Project ID", Value: "could not determine workdir"}
	}
	path := filepath.Join(wd, ".tfplan", "project.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return checkResult{
			Level: "info", Label: "Project ID file:",
			Value: "not present (will be created on first 'plan' run)",
		}
	}
	
	display := strings.TrimSpace(string(data))
	if idx := strings.Index(display, "project_id"); idx >= 0 {
		rest := display[idx:]
		if q := strings.Index(rest, `"`); q >= 0 {
			rest = rest[q+1:]
			if end := strings.Index(rest, `"`); end >= 0 {
				rest = rest[:end]
			}
		}
		display = rest
	}
	if len(display) > 12 {
		display = display[:12]
	}
	return checkResult{Level: "ok", Label: "Project ID file exists:", Value: display}
}

func checkLineages() []checkResult {
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}
	projectID, err := storage.GetOrCreateProjectID(wd)
	if err != nil {
		return nil
	}
	store, err := storage.NewLineageStore(projectID, wd)
	if err != nil {
		return nil
	}
	lineages, err := store.ListLineages()
	if err != nil {
		return []checkResult{{
			Level: "warn", Label: "Lineages",
			Value: fmt.Sprintf("error reading: %v", err),
		}}
	}
	if len(lineages) == 0 {
		return []checkResult{{
			Level: "info", Label: "Lineages found:",
			Value: "0 (run 'plan <name>' to create one)",
		}}
	}
	names := make([]string, 0, len(lineages))
	for _, l := range lineages {
		names = append(names, l.Name)
	}
	return []checkResult{{
		Level: "ok", Label: "Lineages found:",
		Value: fmt.Sprintf("%d (%s)", len(lineages), strings.Join(names, ", ")),
	}}
}

func checkBrokenLineages() []checkResult {
	wd, err := os.Getwd()
	if err != nil {
		return nil
	}
	projectID, err := storage.GetOrCreateProjectID(wd)
	if err != nil {
		return nil
	}
	root := filepath.Join(storageHome(), "projects", projectID)
	entries, err := os.ReadDir(root)
	if err != nil {
		
		return []checkResult{{Level: "ok", Label: "No broken lineages", Value: ""}}
	}

	var hints []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		seen := map[string]int{}
		for _, f := range files {
			name := f.Name()
			switch {
			case strings.HasSuffix(name, ".tfplan"):
				seen[strings.TrimSuffix(name, ".tfplan")] |= 1
			case strings.HasSuffix(name, ".json"):
				seen[strings.TrimSuffix(name, ".json")] |= 2
			}
		}
		for base, flags := range seen {
			if flags == 1 {
				hints = append(hints, fmt.Sprintf("%s/%s.tfplan has no matching .json", e.Name(), base))
			} else if flags == 2 {
				hints = append(hints, fmt.Sprintf("%s/%s.json has no matching .tfplan", e.Name(), base))
			}
		}
	}
	if len(hints) == 0 {
		return []checkResult{{Level: "ok", Label: "No broken lineages", Value: ""}}
	}
	return []checkResult{{
		Level: "warn", Label: "Broken lineages detected:",
		Value: fmt.Sprintf("%d issue(s)", len(hints)),
		Hints: hints,
	}}
}

func storageHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".tfplan-pretty-output")
}
