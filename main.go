// Package main is the entry point for tfplan-pretty-output.
//
// tfplan-pretty-output is a terminal-native Terraform plan reviewer.
// Plan once. Review forever.
package main

import (
	"fmt"
	"os"

	"github.com/OrionVesper/tfplan-pretty-output/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
