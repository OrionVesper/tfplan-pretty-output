package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/OrionVesper/tfplan-pretty-output/storage"
	"github.com/spf13/cobra"
)

var (
	removeYesFlag bool
	removeAllFlag bool
)

var removeCmd = &cobra.Command{
	Use:   "remove [name | name-N]",
	Short: "Delete saved plans (specific, lineage, or all)",
	Long: `Delete saved plans from this project.

Modes:
  remove <name>        Delete the entire lineage (all plans + workdir symlink)
  remove <name>-<N>    Delete a single specific plan
  remove --all         Delete EVERY lineage in this project

By default each remove asks for confirmation. Use -y to skip prompts (for
scripts and automation).

Examples:
  tfplan-pretty-output remove prod-2
  tfplan-pretty-output remove prod
  tfplan-pretty-output remove --all
  tfplan-pretty-output remove prod -y
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&removeYesFlag, "yes", "y", false,
		"Skip confirmation prompts")
	removeCmd.Flags().BoolVar(&removeAllFlag, "all", false,
		"Remove ALL lineages in this project")
}

func runRemove(cmd *cobra.Command, args []string) error {
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

	if removeAllFlag {
		return removeAll(store)
	}

	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return fmt.Errorf(
			"specify a plan name (e.g. 'prod'), a plan-id (e.g. 'prod-2'), or --all\n" +
				"  Run 'tfplan-pretty-output help' for usage")
	}

	ref := args[0]

	if isSpecificPlanID(ref) {
		return removeSpecificPlan(store, ref)
	}

	return removeLineage(store, ref)
}

func removeAll(store *storage.LineageStore) error {
	lineages, err := store.ListLineages()
	if err != nil {
		return err
	}
	if len(lineages) == 0 {
		fmt.Println("No saved plans found for this project — nothing to remove.")
		return nil
	}

	totalPlans := 0
	for _, l := range lineages {
		totalPlans += len(l.Plans)
	}

	if !removeYesFlag {
		msg := fmt.Sprintf(
			"Remove ALL plans for this project (%d lineages, %d plans)? Type 'yes' to confirm: ",
			len(lineages), totalPlans)
		if !promptExactly(msg, "yes") {
			fmt.Println("Aborted.")
			return nil
		}
	}

	for _, l := range lineages {
		if err := store.DeleteLineage(l.Name); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not remove lineage %q: %v\n", l.Name, err)
		}
	}
	fmt.Println("✓ Removed all lineages and symlinks for this project")
	return nil
}

func removeLineage(store *storage.LineageStore, name string) error {
	lineages, err := store.ListLineages()
	if err != nil {
		return err
	}
	var target *storage.Lineage
	for i := range lineages {
		if lineages[i].Name == name {
			target = &lineages[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("lineage %q not found in this project", name)
	}

	if !removeYesFlag {
		msg := fmt.Sprintf(
			"Remove entire lineage %q (%d plans + workdir symlink)? [y/N]: ",
			name, len(target.Plans))
		if !promptYesNo(msg) {
			fmt.Println("Aborted.")
			return nil
		}
	}

	if err := store.DeleteLineage(name); err != nil {
		return err
	}
	fmt.Printf("✓ Removed lineage: %s\n", name)
	return nil
}

func removeSpecificPlan(store *storage.LineageStore, ref string) error {
	p, err := store.FindPlan(ref)
	if err != nil {
		return err
	}

	if !removeYesFlag {
		msg := fmt.Sprintf("Remove plan %q? [y/N]: ", p.ID)
		if !promptYesNo(msg) {
			fmt.Println("Aborted.")
			return nil
		}
	}

	if err := store.DeleteSpecificPlan(p.Lineage, p.Number); err != nil {
		return err
	}
	fmt.Printf("✓ Removed plan: %s\n", p.ID)
	return nil
}

func promptYesNo(message string) bool {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

func promptExactly(message, expected string) bool {
	fmt.Print(message)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	return strings.TrimSpace(line) == expected
}
