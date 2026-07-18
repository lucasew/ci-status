package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "ci-status",
	Short: "A CLI tool to report CI status to forges",
	// Print errors once from main (stderr). Silence cobra's own Error:/Usage
	// dump so messages are not tripled and --silent can stay quiet.
	// Users still get full help via -h/--help.
	SilenceErrors: true,
	SilenceUsage:  true,
}

func main() {
	if err := Command.Execute(); err != nil {
		// --silent wraps the error so CI can still fail without noise.
		if !isQuietError(err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
