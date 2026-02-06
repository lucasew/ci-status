package main

import (
	"context"
	"fmt"
	"errors"
	"os"

	"github.com/spf13/cobra"
	"ci-status/internal/config"
	reporter "ci-status/internal/errors"
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

// execute orchestrates the core logic of the 'run' command.
//
// Flow:
// 1. Validates the CI environment and initializes the forge client (via initForge).
// 2. Reports a 'pending' status to the forge (e.g., GitHub check run).
// 3. Executes the user-specified command with a timeout context.
// 4. Catches specific errors like timeouts (reporting 'error' status and exiting with 124).
// 5. Reports the final status ('success' or 'failure') based on the command's exit code.
// 6. Exits the process with the command's exit code.
//
// Side Effects:
// - Makes HTTP requests to the forge API.
// - Prints warnings/errors to stderr.
// - Terminates the process using os.Exit (does not return).
func execute(cfg config.Config) error {
	ctx := context.Background()
	client, commit := initForge(cfg.Forge, cfg.Commit, cfg.Silent)
	var err error

	// 3. Set Running Status
	if client != nil && commit != "" {
		if err = client.SetStatus(ctx, forge.StatusOpts{
			Commit:      commit,
			Context:     cfg.ContextName,
			State:       forge.StateRunning,
			Description: cfg.PendingDesc,
			TargetURL:   cfg.URL,
		}); err != nil && !cfg.Silent {
			reporter.Warnf("failed to set pending status: %v", err)
		}
	}

	// 5. Execute Command
	exec := executor.New()
	exitCode, err := exec.Run(ctx, cfg.Timeout, cfg.Command, cfg.Args)

    // Handle timeout specifically
    if err != nil && err.Error() == "command timed out" {
        if client != nil && commit != "" {
            if statusErr := client.SetStatus(ctx, forge.StatusOpts{
                Commit:      commit,
                Context:     cfg.ContextName,
                State:       forge.StateError,
                Description: "Timed out",
                TargetURL:   cfg.URL,
            }); statusErr != nil {
                reporter.Report(statusErr)
            }
        }
        reporter.Report(fmt.Errorf("command timed out"))
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
			reporter.Warnf("failed to set final status: %v", err)
		}
	}

	// 7. Exit
	if err != nil && exitCode == 0 {
		// If there was an error running the command but not exit code (e.g. start failed)
		reporter.Report(err)
		os.Exit(1)
	}
	os.Exit(exitCode)
	return nil
}
