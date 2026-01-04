package main

import (
	"context"
	"fmt"
	"errors"
	"os"

	"github.com/spf13/cobra"
	"ci-status/internal/config"
	"ci-status/internal/executor"
	"ci-status/internal/forge"
)

var runConfig config.Config

var ErrCommandMissing = errors.New("command missing after --")

var RunCmd = &cobra.Command{
	Use:   "run [context-name] -- [command] [args...]",
	Short: "Run a command and report status to a forge",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runConfig.ContextName = args[0]

        dashIdx := cmd.ArgsLenAtDash()

        if dashIdx == -1 || dashIdx >= len(args) {
        	return ErrCommandMissing
        }

        runConfig.Command = args[dashIdx]
        if len(args) > dashIdx+1 {
            runConfig.Args = args[dashIdx+1:]
        }

		return execute(runConfig)
	},
}

func init() {
	RunCmd.Flags().StringVar(&runConfig.Forge, "forge", "", "Override automatic forge detection")
	RunCmd.Flags().StringVar(&runConfig.Commit, "commit", "", "Override commit SHA")
	RunCmd.Flags().StringVar(&runConfig.PR, "pr", "", "Override pull request number")
	RunCmd.Flags().StringVar(&runConfig.URL, "url", "", "Target URL for details")
	RunCmd.Flags().StringVar(&runConfig.PendingDesc, "pending-desc", "Running...", "Description shown while command is running")
	RunCmd.Flags().StringVar(&runConfig.SuccessDesc, "success-desc", "Passed", "Description shown when command exits with code 0")
	RunCmd.Flags().StringVar(&runConfig.FailureDesc, "failure-desc", "Failed", "Description shown when command exits with non-zero code")
	RunCmd.Flags().DurationVar(&runConfig.Timeout, "timeout", 0, "Maximum time allowed for command execution")
	RunCmd.Flags().BoolVar(&runConfig.Silent, "silent", false, "Suppress output when running in noop mode or on errors")

	Command.AddCommand(RunCmd)
}

func setupForge(cfg config.Config) (forge.ForgeClient, string) {
	if !isCI(cfg.Silent) {
		return nil, ""
	}

	client, err := forge.DetectClient(cfg.Forge)
	if err != nil {
		if !cfg.Silent {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
		return nil, ""
	}

	commit, err := forge.DetectCommit(cfg.Commit)
	if err != nil {
		if !cfg.Silent {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
		return client, ""
	}

	return client, commit
}

func execute(cfg config.Config) error {
	ctx := context.Background()
	client, commit := setupForge(cfg)

	setInitialStatus(ctx, client, commit, cfg)

	exitCode, err := runCommand(ctx, cfg)

	// Handle timeout reporting separately as it has a special state
	if err != nil && err.Error() == "command timed out" {
		fmt.Fprintln(os.Stderr, "Error: command timed out")
		// The error state is set in setFinalStatus, but we need the specific exit code.
		exitCode = 124
	}

	setFinalStatus(ctx, client, commit, exitCode, err, cfg)

	// Handle process exit code
	if err != nil && exitCode == 0 {
		// Generic error running the command (e.g., command not found)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(exitCode)
	return nil // Unreachable
}

func setInitialStatus(ctx context.Context, client forge.ForgeClient, commit string, cfg config.Config) {
	if client == nil || commit == "" {
		return
	}
	err := client.SetStatus(ctx, forge.StatusOpts{
		Commit:      commit,
		Context:     cfg.ContextName,
		State:       forge.StateRunning,
		Description: cfg.PendingDesc,
		TargetURL:   cfg.URL,
	})
	if err != nil && !cfg.Silent {
		fmt.Fprintf(os.Stderr, "Warning: failed to set pending status: %v\n", err)
	}
}

func runCommand(ctx context.Context, cfg config.Config) (int, error) {
	exec := executor.New()
	return exec.Run(ctx, cfg.Timeout, cfg.Command, cfg.Args)
}

func setFinalStatus(ctx context.Context, client forge.ForgeClient, commit string, exitCode int, err error, cfg config.Config) {
	if client == nil || commit == "" {
		return
	}

	var state forge.State
	var desc string

	if exitCode == 0 && err == nil {
		state = forge.StateSuccess
		desc = cfg.SuccessDesc
	} else if err != nil && err.Error() == "command timed out" {
		state = forge.StateError
		desc = "Timed out"
	} else {
		state = forge.StateFailure
		desc = cfg.FailureDesc
	}

	statusErr := client.SetStatus(ctx, forge.StatusOpts{
		Commit:      commit,
		Context:     cfg.ContextName,
		State:       state,
		Description: desc,
		TargetURL:   cfg.URL,
	})
	if statusErr != nil && !cfg.Silent {
		fmt.Fprintf(os.Stderr, "Warning: failed to set final status: %v\n", statusErr)
	}
}
