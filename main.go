package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ci-status/cmd"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ci-status",
		Short: "A CLI tool to report CI status to forges",
	}

	rootCmd.AddCommand(cmd.RunCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
