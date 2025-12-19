package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "ci-status",
	Short: "A CLI tool to report CI status to forges",
}

func main() {
	if err := Command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
