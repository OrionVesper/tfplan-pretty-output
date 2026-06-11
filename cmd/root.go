package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:           "tfplan-pretty-output",
    Short:         "Terminal-native Terraform plan reviewer",
    SilenceUsage:  true,
    SilenceErrors: true,
}

func Execute() error { return rootCmd.Execute() }

func init() {

    rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
        fmt.Fprintln(os.Stderr, "Error: use 'tfplan-pretty-output help' instead of --help / -h")
        os.Exit(1)
    })
    rootCmd.SetHelpCommand(helpCmd)
    rootCmd.CompletionOptions.DisableDefaultCmd = true
    rootCmd.DisableFlagsInUseLine = true
    rootCmd.DisableAutoGenTag = true

    rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
        if hasDanglingDashDash(os.Args) {
            return fmt.Errorf("no options provided after '--'")
        }
        return nil
    }

    rootCmd.AddCommand(newPassthroughCmd("init", "Initialize the Terraform working directory"))
    rootCmd.AddCommand(newPassthroughCmd("validate", "Check Terraform configuration for syntax errors"))
    rootCmd.AddCommand(newPassthroughCmd("fmt", "Rewrite Terraform configuration files to canonical format"))
    rootCmd.AddCommand(newPassthroughCmd("destroy", "Destroy all remote objects managed by the configuration"))
    rootCmd.AddCommand(newPassthroughCmd("force-unlock", "Release a stuck lock on the current workspace"))

    rootCmd.AddCommand(newPassthroughCmd("output", "Show output values from your root module"))
    rootCmd.AddCommand(newPassthroughCmd("refresh", "Update the state to match remote infrastructure"))
    rootCmd.AddCommand(newPassthroughCmd("providers", "Show the providers required for the configuration"))
    rootCmd.AddCommand(versionCmd)

    rootCmd.AddCommand(planCmd)
    rootCmd.AddCommand(applyCmd)
    rootCmd.AddCommand(viewCmd)
    rootCmd.AddCommand(diffCmd)
    rootCmd.AddCommand(showCmd)
    rootCmd.AddCommand(removeCmd)
    rootCmd.AddCommand(doctorCmd)
}
