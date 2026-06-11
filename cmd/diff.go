package cmd

import (
	"fmt"
	"os"
    "strings"

	"github.com/OrionVesper/tfplan-pretty-output/storage"
	"github.com/OrionVesper/tfplan-pretty-output/terraform"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <plan-1> <plan-2>",
	Short: "Compare two saved plans for this project",
	Long: `Compare two saved plan snapshots and print a categorized summary.

Plan references can be lineage names (latest) or specific IDs.

Examples:
  tfplan-pretty-output diff prod prod-98
  tfplan-pretty-output diff prod-97 prod-98
`,
    DisableFlagParsing: true, 
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

func runDiff(cmd *cobra.Command, args []string) error {

    if strings.HasPrefix(args[0], "-") {
        return fmt.Errorf("plan name is required (got flag-like value %q)", args[0])
    }
    if strings.HasPrefix(args[1], "-") {
        return fmt.Errorf("plan name is required (got flag-like value %q)", args[1])
    }

	ref1, ref2 := args[0], args[1]

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

	plan1, err := loadParsedPlan(store, ref1)
	if err != nil {
		return err
	}
	plan2, err := loadParsedPlan(store, ref2)
	if err != nil {
		return err
	}

	d := terraform.ComputePlanDiff(plan1, plan2)
	id1, _ := resolveID(store, ref1)
	id2, _ := resolveID(store, ref2)
	printDiff(id1, id2, d)
	return nil
}

func loadParsedPlan(store *storage.LineageStore, ref string) (*terraform.Plan, error) {
	path, err := store.FindPlanJSON(ref)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read saved plan %s: %w", ref, err)
	}
	return terraform.ParsePlanJSON(data)
}

func resolveID(store *storage.LineageStore, ref string) (string, error) {
	p, err := store.FindPlan(ref)
	if err != nil {
		return ref, err
	}
	return p.ID, nil
}

func printDiff(id1, id2 string, d *terraform.PlanDiff) {
	fmt.Println()
	fmt.Printf("tfplan-pretty-output                              diff: %s → %s\n", id1, id2)
	fmt.Println()

	if len(d.Added) > 0 {
		fmt.Printf("  Added (%d):\n", len(d.Added))
		for _, rc := range d.Added {
			sym := rc.Change.Actions.ActionType().Symbol()
			fmt.Printf("    + %-60s %s\n", rc.Address, sym)
		}
		fmt.Println()
	}
	if len(d.Removed) > 0 {
		fmt.Printf("  Removed (%d):\n", len(d.Removed))
		for _, rc := range d.Removed {
			sym := rc.Change.Actions.ActionType().Symbol()
			fmt.Printf("    - %-60s %s\n", rc.Address, sym)
		}
		fmt.Println()
	}
	if len(d.Changed) > 0 {
		fmt.Printf("  Changed action (%d):\n", len(d.Changed))
		for _, e := range d.Changed {
			fmt.Printf("    ~ %-60s %s → %s\n", e.Address, e.Before.Symbol(), e.After.Symbol())
		}
		fmt.Println()
	}
	fmt.Printf("  Unchanged: %d\n", d.Unchanged)
	fmt.Println()
}
