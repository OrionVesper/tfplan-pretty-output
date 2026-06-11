package cmd

import (
    "fmt"
    "strings"

    "github.com/spf13/cobra"
)

var terraformPassthroughCommands = []string{
    "init", "validate", "fmt", "destroy", "force-unlock",
    "output", "refresh", "providers", "version",
}

var helpCmd = &cobra.Command{
    Use:                "help [command]",
    Short:              "Show available commands and usage",
    DisableFlagParsing: true,
    Args:               cobra.ArbitraryArgs,
    RunE:               runHelp,
}

func runHelp(cmd *cobra.Command, args []string) error {
    if len(args) == 0 {
        printOverview()
        return nil
    }
    if len(args) > 1 {
        return fmt.Errorf("too many arguments — only one command name allowed")
    }

    target := args[0]
    if strings.HasPrefix(target, "-") {
        return fmt.Errorf("command name required (got flag-like value %q)", target)
    }

    return showCommandHelp(target)
}

func showCommandHelp(target string) error {

    switch target {
    case "plan":
        fmt.Print(helpPlan)
        fmt.Println()
        return runTerraformHelp("plan")
    case "apply":
        fmt.Print(helpApply)
        fmt.Println()
        return runTerraformHelp("apply")
    }

    customHelp := map[string]string{
        "view":   helpView,
        "diff":   helpDiff,
        "show":   helpShow,
        "remove": helpRemove,
        "doctor": helpDoctor,
        "help":   helpHelp,
    }
    if text, ok := customHelp[target]; ok {
        fmt.Print(text)
        return nil
    }

    for _, c := range terraformPassthroughCommands {
        if c == target {
            return runTerraformHelp(target)
        }
    }

    return fmt.Errorf(
        "no help available for %q — see 'tfplan-pretty-output help' for the list of commands",
        target)
}

func printOverview() {
    fmt.Println(`tfplan-pretty-output — Terminal-native Terraform plan reviewer

Usage:
  tfplan-pretty-output <command> [arguments]
  tfplan-pretty-output help <command>    # see detailed help for a command

Lifecycle Commands:
  init           Initialize the Terraform working directory
  validate       Check Terraform configuration for syntax errors
  fmt            Rewrite Terraform configuration files to canonical format
  plan           Run terraform plan and save snapshot (requires lineage name)
  apply          Apply the latest plan in a lineage
  destroy        Destroy all remote objects managed by the configuration
  force-unlock   Release a stuck lock on the current workspace

Inspection Commands:
  output      Show output values from your root module
  refresh     Update the state to match remote infrastructure
  providers   Show the providers required for the configuration
  version     Show tfplan-pretty-output and terraform versions

tfplan-pretty-output Commands:
  view        List or open saved plans
  diff        Compare two saved plans
  show        Print a saved plan as text (pipe-friendly)
  remove      Delete saved plans (specific, lineage, or --all)
  doctor      Diagnose common setup issues
  help        Show this help

Saved plans live under:
  ~/.tfplan-pretty-output/projects/<project-id>/<lineage-name>/

Each capture creates a symlink in your working directory pointing to the
latest plan of that lineage (e.g. ./prod).
`)
}