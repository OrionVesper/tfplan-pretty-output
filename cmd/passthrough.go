package cmd

import (
    "os"

    "github.com/OrionVesper/tfplan-pretty-output/terraform"
    "github.com/spf13/cobra"
)

func newPassthroughCmd(name, short string) *cobra.Command {
    return &cobra.Command{
        Use:   name + " [terraform flags...]",
        Short: short,
        Long: short + `.

This command is a transparent wrapper around 'terraform ` + name + `'.
All flags are passed through directly to terraform, so any flag supported
by 'terraform ` + name + `' works here unchanged.

For the authoritative list of flags, run:

  terraform ` + name + ` -help`,

        DisableFlagParsing: true,

        RunE: func(cmd *cobra.Command, args []string) error {
            runner, err := terraform.NewRunner(terraform.RunnerOptions{})
            if err != nil {
                return err
            }

            err = runner.Passthrough(name, os.Stdout, args)
            if err == nil {
                return nil
            }

            if terraform.IsInterrupted(err) {
                terraform.PrintInterruptMessage("")
                os.Exit(130)
            }

            if exitErr, ok := terraform.IsExitError(err); ok {
                os.Exit(exitErr.ExitCode)
            }

            return err
        },
    }
}