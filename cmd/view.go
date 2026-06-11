package cmd

import (
	"fmt"
	"os"
	"sort"
    "strings"
	
	"github.com/OrionVesper/tfplan-pretty-output/storage"
	"github.com/OrionVesper/tfplan-pretty-output/terraform"
	"github.com/OrionVesper/tfplan-pretty-output/tui"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view <name | name-N | list>",
	Short: "Open a saved plan, a specific plan, or list saved plans",
	Long: `Open the latest plan of a lineage, a specific plan within a lineage,
or list all saved plans for this project.

Examples:
  tfplan-pretty-output view prod
  tfplan-pretty-output view prod-98
  tfplan-pretty-output view list
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runView,
}

func runView(cmd *cobra.Command, args []string) error {
	if len(args) == 1 && args[0] == "list" {
		return printSavedPlans()
	}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return fmt.Errorf(
			"plan name is required\n  Usage: tfplan-pretty-output view <name|name-N|list>")
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
	jsonPath, err := store.FindPlanJSON(ref)
	if err != nil {
		return err
	}

	if info, err := os.Stat(planFilePath); err != nil || info.Size() == 0 {
		return fmt.Errorf(
			"plan file is missing or empty for %s\n  → try running: tfplan-pretty-output plan %s",
			ref, stripTrailingNumber(ref))
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("could not read saved plan %s: %w", ref, err)
	}
	plan, err := terraform.ParsePlanJSON(data)
	if err != nil {
		return fmt.Errorf("could not parse saved plan %s: %w", ref, err)
	}

	if tui.IsEmptyPlan(plan) {
		fmt.Printf("✓ Plan: %s\n\n", ref)
		fmt.Println("  No resource changes — nothing to view.")
		fmt.Println("  This plan only reads data sources or has no actionable changes.")
		return nil
	}

	p, err := store.FindPlan(ref)
	if err != nil {
		return err
	}
	return tui.Run(p.ID, plan, planFilePath, workdir)
}

func printSavedPlans() error {
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

	lineages, err := store.ListLineages()
	if err != nil {
		return err
	}
	if len(lineages) == 0 {
		fmt.Println("No saved plans yet for this directory.")
		fmt.Println()
		fmt.Println("Run:")
		fmt.Println("  tfplan-pretty-output plan <name>")
		return nil
	}

	sort.SliceStable(lineages, func(i, j int) bool {
		return lineages[i].LatestStamp.After(lineages[j].LatestStamp)
	})

    fmt.Println("saved plans                                                            tfplan-pretty-output")
    fmt.Println()

	for li, l := range lineages {
		fmt.Printf("  %s\n", l.Name)
		for pi, p := range l.Plans {
			marker := ""
			if pi == 0 {
				if li == 0 {
					marker = "  ← latest"
				} else {
					marker = "  ← latest old"
				}
			}
			ts := p.Timestamp.Local().Format("2006-01-02 15:04")
			fmt.Printf("    %-32s  %-16s%s\n", p.ID, ts, marker)
		}
		fmt.Println()
	}

	fmt.Println("  (Use 'tfplan-pretty-output view <name>' to open the latest plan in a lineage.)")
	return nil
}
