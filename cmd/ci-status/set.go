package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ci-status/internal/forge"
)

type SetConfig struct {
	ContextName string
	State       string
	Description string
	URL         string
	Commit      string
	PR          string
	Forge       string
	Silent      bool
}

var setConfig SetConfig

var SetCmd = &cobra.Command{
	Use:   "set [context-name]",
	Short: "Set a status for a context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		setConfig.ContextName = args[0]
		return executeSet(setConfig)
	},
}

func init() {
	SetCmd.Flags().StringVar(&setConfig.State, "state", "pending", "State to set (pending, success, failure, error)")
	SetCmd.Flags().StringVar(&setConfig.Description, "description", "", "Description of the status")
	SetCmd.Flags().StringVar(&setConfig.URL, "url", "", "Target URL")
	SetCmd.Flags().StringVar(&setConfig.Commit, "commit", "", "Override commit SHA")
	SetCmd.Flags().StringVar(&setConfig.PR, "pr", "", "Override pull request number")
	SetCmd.Flags().StringVar(&setConfig.Forge, "forge", "", "Override automatic forge detection")
	SetCmd.Flags().BoolVar(&setConfig.Silent, "silent", false, "Suppress output")

	Command.AddCommand(SetCmd)
}

func executeSet(cfg SetConfig) error {
	ctx := context.Background()

	// 1. Detect Forge
	forgeName, err := forge.DetectForge(cfg.Forge)
	if err != nil && !cfg.Silent {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// 2. Detect Commit
	commit, err := forge.DetectCommit(cfg.Commit)
	if err != nil && !cfg.Silent {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// 3. Initialize Client
	var client forge.ForgeClient
	if forgeName == "github" {
		token := os.Getenv("GITHUB_TOKEN")
		if token != "" {
			owner, repo, err := forge.DetectRepoInfo()
			if err == nil {
				client = forge.NewGitHubClient(token, owner, repo)
			} else if !cfg.Silent {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}
		} else if !cfg.Silent {
			fmt.Fprintf(os.Stderr, "Warning: GITHUB_TOKEN not set\n")
		}
	}

	// 4. Set Status
	if client != nil && commit != "" {
		err := client.SetStatus(ctx, forge.StatusOpts{
			Commit:      commit,
			Context:     cfg.ContextName,
			State:       forge.State(cfg.State),
			Description: cfg.Description,
			TargetURL:   cfg.URL,
		})
		if err != nil {
			if !cfg.Silent {
				fmt.Fprintf(os.Stderr, "Error: failed to set status: %v\n", err)
			}
			return err
		}
	} else {
        if !cfg.Silent {
            fmt.Fprintln(os.Stderr, "Noop: Forge client or commit not available")
        }
    }

	return nil
}
