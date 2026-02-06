package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	reporter "ci-status/internal/errors"
	"ci-status/internal/forge"
)

// SetConfig holds the configuration for the 'set' command, which manually
// reports a specific status without running a wrapped command.
type SetConfig struct {
	// ContextName is the status identifier (e.g., "lint", "deploy").
	ContextName string
	// State is the target status (pending, success, failure, error).
	State string
	// Description is a short text explaining the status.
	Description string
	// URL provides a link to further details (e.g., logs).
	URL string
	// Commit overrides the detected commit SHA.
	Commit string
	// PR overrides the detected PR number.
	PR string
	// Forge overrides the detected forge type.
	Forge string
	// Silent suppresses warning/error messages.
	Silent bool
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

	if !isCI(cfg.Silent) {
		return nil
	}

	// 1. Detect Forge Client
	client, err := forge.DetectClient(cfg.Forge)
	if err != nil {
		if !cfg.Silent {
			reporter.Warn(err)
		}
		client = nil
	}

	// 2. Detect Commit
	commit, err := forge.DetectCommit(cfg.Commit)
	if err != nil && !cfg.Silent {
		reporter.Warn(err)
	}

	// 3. Set Status
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
				reporter.Report(fmt.Errorf("failed to set status: %w", err))
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
