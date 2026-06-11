package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/OrionVesper/tfplan-pretty-output/storage"
	"github.com/OrionVesper/tfplan-pretty-output/terraform"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply <lineage-name>",
	Short: "Apply the latest plan in a lineage",
	Long: `Apply the latest plan saved in the given lineage.

Examples:
  tfplan-pretty-output apply prod
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runApply,
}

func runApply(cmd *cobra.Command, args []string) error {
    if len(args) == 0 || strings.HasPrefix(args[0], "-") {
        return fmt.Errorf(
            "plan name is required\n  Usage: tfplan-pretty-output apply <name>")
    }

	ref := args[0]

	if isSpecificPlanID(ref) {
		fmt.Fprintln(os.Stderr, "Error: applying a specific plan is not supported.")
		fmt.Fprintln(os.Stderr, "")
		lineage := stripTrailingNumber(ref)
		fmt.Fprintf(os.Stderr, "  Use 'tfplan-pretty-output apply %s' to apply the latest plan.\n", lineage)
		fmt.Fprintln(os.Stderr, "  Old plans can drift from current code and corrupt the state.")
		os.Exit(1)
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

	planFilePath, err := store.FindPlanFile(ref)
	if err != nil {
		return err
	}

    c := exec.Command("terraform", "apply", planFilePath)
    c.Dir = workdir
    c.Env = os.Environ()
    c.Stdin = os.Stdin
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    
    err = terraform.RunWithSignalHandling(c)
    if err != nil {
        if terraform.IsInterrupted(err) {
            terraform.PrintInterruptMessage("")
            os.Exit(130)
        }
        return err
    }
    return nil
}

func isSpecificPlanID(s string) bool {
	idx := strings.LastIndex(s, "-")
	if idx <= 0 || idx == len(s)-1 {
		return false
	}
	suffix := s[idx+1:]
	for _, c := range suffix {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func stripTrailingNumber(s string) string {
	idx := strings.LastIndex(s, "-")
	if idx <= 0 {
		return s
	}
	return s[:idx]
}
