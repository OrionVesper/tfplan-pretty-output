package cmd

import (
    "bytes"
    "fmt"
    "os"
    "os/exec"

    "github.com/OrionVesper/tfplan-pretty-output/terraform"
)

func runTerraformHelp(subcommand string) error {
    if _, err := exec.LookPath("terraform"); err != nil {
        return fmt.Errorf("terraform binary not found on PATH")
    }

    var stdout, stderr bytes.Buffer
    c := exec.Command("terraform", subcommand, "-help")
    c.Stdout = &stdout
    c.Stderr = &stderr
    _ = c.Run()

    output := stdout.String() + stderr.String()

    rewriter := terraform.NewCommandRewriter(os.Stdout)
    if _, err := rewriter.Write([]byte(output)); err != nil {
        return err
    }
    if f, ok := rewriter.(interface{ Flush() error }); ok {
        _ = f.Flush()
    }
    return nil
}