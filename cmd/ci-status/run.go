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

func execute(cfg config.Config) error {
	ctx := context.Background()

	// 1. Detect Forge Client
	// This replaces the previous separate steps for Forge Name -> Repo Info -> New Client
	client, err := forge.DetectClient(cfg.Forge)
	if err != nil {
		if !cfg.Silent {
			// It's a warning, not necessarily a fatal error if we want noop mode?
			// But DetectClient returns error if no supported forge found.
			// Previous logic printed Warning and proceeded (client=nil).
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
		client = nil
	}

	// 2. Detect Commit
	commit, err := forge.DetectCommit(cfg.Commit)
	if err != nil && !cfg.Silent {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// 3. Set Running Status
	if client != nil && commit != "" {
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

	// 5. Execute Command
	exec := executor.New()
	exitCode, err := exec.Run(ctx, cfg.Timeout, cfg.Command, cfg.Args)

    // Handle timeout specifically
    if err != nil && err.Error() == "command timed out" {
        if client != nil && commit != "" {
            _ = client.SetStatus(ctx, forge.StatusOpts{
                Commit:      commit,
                Context:     cfg.ContextName,
                State:       forge.StateError,
                Description: "Timed out",
                TargetURL:   cfg.URL,
            })
        }
        fmt.Fprintln(os.Stderr, "Error: command timed out")
        os.Exit(124)
    }

	// 6. Set Final Status
	if client != nil && commit != "" {
		var state forge.State
		var desc string

		if exitCode == 0 && err == nil {
			state = forge.StateSuccess
			desc = cfg.SuccessDesc
		} else {
			state = forge.StateFailure
			desc = cfg.FailureDesc
		}

		err := client.SetStatus(ctx, forge.StatusOpts{
			Commit:      commit,
			Context:     cfg.ContextName,
			State:       state,
			Description: desc,
			TargetURL:   cfg.URL,
		})
		if err != nil && !cfg.Silent {
			fmt.Fprintf(os.Stderr, "Warning: failed to set final status: %v\n", err)
		}
	}

	// 7. Exit
	if err != nil && exitCode == 0 {
		// If there was an error running the command but not exit code (e.g. start failed)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
	return nil
}
