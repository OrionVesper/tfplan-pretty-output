package cmd

import (
	"fmt"
	"os"
    "strings"

	"github.com/OrionVesper/tfplan-pretty-output/storage"
	"github.com/OrionVesper/tfplan-pretty-output/terraform"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <name | name-N>",
	Short: "Print a saved plan as text (pipe-friendly)",
	Long: `Re-runs 'terraform show' on a saved plan and prints the full text.

Examples:
  tfplan-pretty-output show prod
  tfplan-pretty-output show prod-98 | less
`,
    DisableFlagParsing: true, 
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func runShow(cmd *cobra.Command, args []string) error {

    if strings.HasPrefix(args[0], "-") {
        return fmt.Errorf("plan name is required (got flag-like value %q)", args[0])
    }

	ref := args[0]

	workdir, err := os.Getwd()
	if err != nil {
		return err
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
	if info, err := os.Stat(planFilePath); err != nil || info.Size() == 0 {
		return fmt.Errorf("plan file is missing or empty for %s", ref)
	}

	text, err := terraform.ShowPlanText(workdir, planFilePath)
	if err != nil {
		return err
	}
	fmt.Print(text)
	return nil
}
