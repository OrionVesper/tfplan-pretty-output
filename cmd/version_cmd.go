package cmd

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Show tfplan-pretty-output and terraform versions",
    Long: `Show the version of tfplan-pretty-output and the underlying
terraform binary it wraps.`,
    DisableFlagParsing: true,
    RunE: func(cmd *cobra.Command, args []string) error {

        fmt.Printf("tfplan-pretty-output %s\n", Version)
        if Commit != "none" {
            fmt.Printf("  commit: %s\n", Commit)
        }
        if Date != "unknown" {
            fmt.Printf("  built:  %s\n", Date)
        }

        fmt.Println()

        c := exec.Command("terraform", "version")
        c.Stdout = os.Stdout
        c.Stderr = os.Stderr
        _ = c.Run()

        return nil
    },
}