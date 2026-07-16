package main

import (
	"context"
	"fmt"
	"os"

	"ci-status/internal/forge"
	"github.com/spf13/cobra"
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
	SetCmd.Flags().StringVar(&setConfig.State, "state", "pending", "State to set (pending, success, failure, error, running)")
	SetCmd.Flags().StringVar(&setConfig.Description, "description", "", "Description of the status")
	SetCmd.Flags().StringVar(&setConfig.URL, "url", "", "Target URL")
	SetCmd.Flags().StringVar(&setConfig.Commit, "commit", "", "Override commit SHA")
	SetCmd.Flags().StringVar(&setConfig.PR, "pr", "", "Override pull request number")
	SetCmd.Flags().StringVar(&setConfig.Forge, "forge", "", "Override automatic forge detection")
	SetCmd.Flags().BoolVar(&setConfig.Silent, "silent", false, "Suppress output")

	Command.AddCommand(SetCmd)
}

// parseState maps a CLI --state value to a forge.State.
// Unknown values fail early so the forge API never receives invalid statuses.
func parseState(s string) (forge.State, error) {
	state := forge.State(s)
	switch state {
	case forge.StatePending, forge.StateSuccess, forge.StateFailure, forge.StateError, forge.StateRunning:
		return state, nil
	default:
		return "", fmt.Errorf("invalid state %q (want pending|success|failure|error|running)", s)
	}
}

func executeSet(cfg SetConfig) error {
	ctx := context.Background()

	state, err := parseState(cfg.State)
	if err != nil {
		return err
	}

	// Outside CI, skip reporting (same policy as run). Unlike run, set's only
	// job is to post a status — once we are in CI, failures must be errors so
	// scripts do not treat a missed status as success.
	if !isCI(cfg.Silent) {
		return nil
	}

	client, err := forge.DetectClient(cfg.Forge)
	if err != nil {
		if !cfg.Silent {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return err
	}

	commit, err := forge.DetectCommit(cfg.Commit)
	if err != nil {
		if !cfg.Silent {
			fmt.Fprintf(os.Stderr, "Error: commit not available: %v\n", err)
		}
		return fmt.Errorf("commit not available: %w", err)
	}
	if commit == "" {
		err := fmt.Errorf("commit not available")
		if !cfg.Silent {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return err
	}

	if err := client.SetStatus(ctx, forge.StatusOpts{
		Commit:      commit,
		Context:     cfg.ContextName,
		State:       state,
		Description: cfg.Description,
		TargetURL:   cfg.URL,
	}); err != nil {
		if !cfg.Silent {
			fmt.Fprintf(os.Stderr, "Error: failed to set status: %v\n", err)
		}
		return err
	}

	return nil
}
