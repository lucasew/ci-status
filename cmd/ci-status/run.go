package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"ci-status/internal/config"
	"ci-status/internal/executor"
	"ci-status/internal/forge"
	"github.com/spf13/cobra"
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

		return execute(cmd.Context(), runConfig)
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

// postStatus reports a forge status when a client and commit are available.
// API failures are warnings only (unless silent); they must not fail the run.
// label is the human phrase in the warning ("pending", "timeout", "final").
func postStatus(ctx context.Context, client forge.ForgeClient, commit string, silent bool, opts forge.StatusOpts, label string) {
	if client == nil || commit == "" {
		return
	}
	if err := client.SetStatus(ctx, opts); err != nil && !silent {
		fmt.Fprintf(os.Stderr, "Warning: failed to set %s status: %v\n", label, err)
	}
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
//
// ctx should come from the cobra command (cmd.Context()) so a parent
// ExecuteContext cancel reaches status posts and the wrapped command.
func execute(ctx context.Context, cfg config.Config) error {
	client, commit := initForge(cfg.Forge, cfg.Commit, cfg.Silent)

	// Shared StatusOpts fields for every post in this run.
	base := forge.StatusOpts{
		Commit:    commit,
		Context:   cfg.ContextName,
		TargetURL: cfg.URL,
	}

	// 3. Set Running Status
	pending := base
	pending.State = forge.StateRunning
	pending.Description = cfg.PendingDesc
	postStatus(ctx, client, commit, cfg.Silent, pending, "pending")

	// 5. Execute Command
	exec := executor.New()
	exitCode, err := exec.Run(ctx, cfg.Timeout, cfg.Command, cfg.Args)

	// Handle timeout specifically
	if errors.Is(err, executor.ErrTimeout) {
		timeoutOpts := base
		timeoutOpts.State = forge.StateError
		timeoutOpts.Description = "Timed out"
		postStatus(ctx, client, commit, cfg.Silent, timeoutOpts, "timeout")
		// Match final/start paths and --silent ("on errors"): still exit 124.
		if !cfg.Silent {
			fmt.Fprintln(os.Stderr, "Error: command timed out")
		}
		os.Exit(executor.ExitCodeTimeout)
	}

	// 6. Set Final Status — do not shadow executor err: start failures return
	// exitCode 0 with a non-nil error, and the exit path below must still see it.
	state, desc := finalStatus(exitCode, err, cfg.SuccessDesc, cfg.FailureDesc)
	finalOpts := base
	finalOpts.State = state
	finalOpts.Description = desc
	postStatus(ctx, client, commit, cfg.Silent, finalOpts, "final")

	// 7. Exit
	if err != nil && exitCode == 0 {
		// Start failed (e.g. executable not found). Exit 1; respect --silent.
		if !cfg.Silent {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
	os.Exit(exitCode)
	return nil
}

// finalStatus maps an executor result to the forge status that should be posted.
//
// executor.Run returns exitCode 0 with a non-nil error when the process never
// started (e.g. executable not found). That is a runtime/config problem
// (StateError), not a failed check (StateFailure, reserved for real exit codes).
func finalStatus(exitCode int, err error, successDesc, failureDesc string) (forge.State, string) {
	if exitCode == 0 && err == nil {
		return forge.StateSuccess, successDesc
	}
	if err != nil && exitCode == 0 {
		return forge.StateError, "Failed to start"
	}
	return forge.StateFailure, failureDesc
}
