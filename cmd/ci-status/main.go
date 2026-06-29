package main

import (
	"os"

	"ci-status/internal/reporter"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "ci-status",
	Short: "A CLI tool to report CI status to forges",
}

func main() {
	if err := Command.Execute(); err != nil {
		reporter.ReportError(err)
		os.Exit(1)
	}
}
