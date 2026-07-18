package main

import (
	"errors"
	"testing"

	"ci-status/internal/forge"
)

func TestFinalStatus(t *testing.T) {
	tests := []struct {
		name        string
		exitCode    int
		err         error
		wantState   forge.State
		wantDesc    string
	}{
		{
			name:      "success",
			exitCode:  0,
			err:       nil,
			wantState: forge.StateSuccess,
			wantDesc:  "Passed",
		},
		{
			name:      "non-zero exit is failure",
			exitCode:  2,
			err:       nil,
			wantState: forge.StateFailure,
			wantDesc:  "Failed",
		},
		{
			name:      "start failure is error not failure",
			exitCode:  0,
			err:       errors.New("failed to start command: exec: \"nope\": executable file not found"),
			wantState: forge.StateError,
			wantDesc:  "Failed to start",
		},
		{
			// Timeout is handled before finalStatus is called, but if it were
			// passed through with a non-zero code, treat as failure (exit code wins).
			name:      "non-zero with error still failure",
			exitCode:  124,
			err:       errors.New("command timed out"),
			wantState: forge.StateFailure,
			wantDesc:  "Failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, desc := finalStatus(tt.exitCode, tt.err, "Passed", "Failed")
			if state != tt.wantState {
				t.Fatalf("state = %q, want %q", state, tt.wantState)
			}
			if desc != tt.wantDesc {
				t.Fatalf("desc = %q, want %q", desc, tt.wantDesc)
			}
		})
	}
}
