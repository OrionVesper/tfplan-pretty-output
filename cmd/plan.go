package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/OrionVesper/tfplan-pretty-output/storage"
	"github.com/OrionVesper/tfplan-pretty-output/terraform"
	"github.com/OrionVesper/tfplan-pretty-output/tui"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:                "plan <name> [-flags...] [-- view | auto-clean]",
	Short:              "Run tfplan-pretty-output plan and save a snapshot in the named lineage",
	Long: `Capture a tfplan-pretty-output plan as a snapshot in a named lineage.

Examples:
  tfplan-pretty-output plan prod
  tfplan-pretty-output plan prod -refresh-only
  tfplan-pretty-output plan prod -- view
  tfplan-pretty-output plan prod -- auto-clean
`,
	DisableFlagParsing: true,
	Args:               cobra.ArbitraryArgs,
	RunE:               runPlan,
}

var lineageNameRe = regexp.MustCompile(`^[a-zA-Z]+$`)

func runPlan(cmd *cobra.Command, args []string) error {
    if len(args) == 0 || strings.HasPrefix(args[0], "-") {
        return fmt.Errorf(
            "plan name is required\n  Usage: tfplan-pretty-output plan <name> [-flags...]")
    }

    name := args[0]
    if err := validateLineageName(name); err != nil {
        return err
    }

	tfArgs, ourArgs := splitAtDashDash(args[1:])

	opts, err := parsePlanOptions(ourArgs)
	if err != nil {
		return err
	}

	if hasOutFlag(tfArgs) {
		return fmt.Errorf(
			"the -out flag is not allowed; tfplan-pretty-output manages plan files automatically")
	}

	workdir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	projectID, err := storage.GetOrCreateProjectID(workdir)
	if err != nil {
		return err
	}
	store, err := storage.NewLineageStore(projectID, workdir)
	if err != nil {
		return err
	}

	tmpPlanPath, planJSON, planObj, err := runTerraformPlanAndCaptureJSON(workdir, tfArgs, os.Stdout)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			_ = os.Remove(tmpPlanPath)
			os.Exit(exitErr.ExitCode())
		}
		_ = os.Remove(tmpPlanPath)
		return err
	}

	if opts.AutoClean {
		_ = os.Remove(tmpPlanPath)
		fmt.Println()
		fmt.Println("✓ Plan completed (not saved — auto-clean mode)")
		return nil
	}

	id, err := store.SavePlan(name, planObj, planJSON, tmpPlanPath)
	if err != nil {
		return fmt.Errorf("could not save plan snapshot: %w", err)
	}

	fmt.Printf("\n✓ Plan captured: %s\n\n", id)
	fmt.Println("To view, run:")
	fmt.Printf("  tfplan-pretty-output view %s\n\n", name)
	fmt.Println("To apply, run:")
	fmt.Printf("  tfplan-pretty-output apply %s\n", name)

	if opts.View {
		if tui.IsEmptyPlan(planObj) {
			fmt.Println()
			fmt.Println("  No resource changes — nothing to view.")
			return nil
		}
		planFilePath, ferr := store.FindPlanFile(name)
		if ferr != nil {
			return fmt.Errorf("could not locate plan file for viewing: %w", ferr)
		}
		if err := tui.Run(id, planObj, planFilePath, workdir); err != nil {
			return fmt.Errorf("could not open viewer: %w", err)
		}
	}
	return nil
}

func validateLineageName(name string) error {
    if name == "" {
        return fmt.Errorf("lineage name cannot be empty")
    }

    if !lineageNameRegex.MatchString(name) {
        return fmt.Errorf(
            "invalid lineage name %q: only letters (a-z, A-Z) are allowed; "+
                "no numbers, hyphens, symbols, or spaces",
            name)
    }

    return nil
}

func hasOutFlag(args []string) bool {
	for _, a := range args {
		if a == "-out" || strings.HasPrefix(a, "-out=") {
			return true
		}
	}
	return false
}

func splitAtDashDash(args []string) (tfArgs, ourArgs []string) {
	for i, a := range args {
		if a == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

type planOptions struct {
	View      bool
	AutoClean bool
}

func parsePlanOptions(args []string) (planOptions, error) {
	var o planOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "view":
			o.View = true
		case "auto-clean":
			o.AutoClean = true
		case "":
			// ignore
		default:
			return o, fmt.Errorf("unknown option: %s", args[i])
		}
	}
	if o.View && o.AutoClean {
		return o, fmt.Errorf("'view' and 'auto-clean' cannot be used together")
	}
	return o, nil
}

func runTerraformPlanAndCaptureJSON(
	workdir string, tfArgs []string, userOut io.Writer,
) (string, []byte, *terraform.Plan, error) {
	tmp, err := os.CreateTemp("", "tfplan-pretty-output-*.tfplan")
	if err != nil {
		return "", nil, nil, fmt.Errorf("could not create temp plan file: %w", err)
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	_ = os.Remove(tmpPath) // terraform will recreate

	filter := newPlanOutputFilter(userOut)
	rewriter := terraform.NewCommandRewriter(filter)

	args := append([]string{"plan", "-out=" + tmpPath}, tfArgs...)
    c := exec.Command("terraform", args...)
    c.Dir = workdir
    c.Env = os.Environ()
    c.Stdin = os.Stdin
    c.Stdout = rewriter
    c.Stderr = rewriter
    
    err = terraform.RunWithSignalHandling(c)
    if err != nil {
        if terraform.IsInterrupted(err) {
            terraform.PrintInterruptMessage("plan")
            _ = os.Remove(tmpPath)
            os.Exit(130)
        }
        return tmpPath, nil, nil, err
    }
	if f, ok := rewriter.(interface{ Flush() error }); ok {
		_ = f.Flush()
	}
	_ = filter.Flush()

	var jsonBuf bytes.Buffer
    show := exec.Command("terraform", "show", "-json", tmpPath)
    show.Dir = workdir
    show.Env = os.Environ()
    show.Stdin = os.Stdin
    show.Stdout = &jsonBuf
    show.Stderr = userOut
    
    err = terraform.RunWithSignalHandling(show)
    if err != nil {
        if terraform.IsInterrupted(err) {
            terraform.PrintInterruptMessage("")
            _ = os.Remove(tmpPath)
            os.Exit(130)
        }
        return tmpPath, nil, nil, fmt.Errorf("terraform show -json failed: %w", err)
    }

	jsonBytes := jsonBuf.Bytes()
	planObj, err := terraform.ParsePlanJSON(jsonBytes)
	if err != nil {
		return tmpPath, nil, nil, err
	}
	return tmpPath, jsonBytes, planObj, nil
}

type planOutputFilter struct {
	out  io.Writer
	buf  bytes.Buffer
	skip bool
}

func newPlanOutputFilter(w io.Writer) *planOutputFilter {
	return &planOutputFilter{out: w}
}

func (f *planOutputFilter) Write(p []byte) (int, error) {
	f.buf.Write(p)
	for {
		idx := bytes.IndexByte(f.buf.Bytes(), '\n')
		if idx < 0 {
			break
		}
		line := string(f.buf.Next(idx + 1))
		trim := strings.TrimSpace(line)

		if strings.Contains(line, "Saved the plan to:") {
			f.skip = true
			continue
		}
		if strings.Contains(line, "To perform exactly these actions") {
			f.skip = true
			continue
		}
		if f.skip && trim == "" {
			f.skip = false
			continue
		}
		if f.skip {
			continue
		}

		if _, err := io.WriteString(f.out, line); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (f *planOutputFilter) Flush() error {
	if f.buf.Len() == 0 {
		return nil
	}
	rest := string(f.buf.Bytes())
	f.buf.Reset()
	if f.skip {
		return nil
	}
	_, err := io.WriteString(f.out, rest)
	return err
}
