package executor

import (
	"context"
	"errors"
	"testing"
)

func TestOutcomeAfterStop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		waitErr  error
		ctxErr   error
		wantCode int
		wantErr  error // errors.Is target; nil means want nil err
	}{
		{
			// Both select cases ready and Wait already saw exit 0: must not
			// report timeout/cancel (the bug this helper fixes).
			name:     "success wins over deadline",
			waitErr:  nil,
			ctxErr:   context.DeadlineExceeded,
			wantCode: 0,
			wantErr:  nil,
		},
		{
			name:     "success wins over cancel",
			waitErr:  nil,
			ctxErr:   context.Canceled,
			wantCode: 0,
			wantErr:  nil,
		},
		{
			name:     "timeout when process did not succeed",
			waitErr:  errors.New("signal: killed"),
			ctxErr:   context.DeadlineExceeded,
			wantCode: ExitCodeTimeout,
			wantErr:  ErrTimeout,
		},
		{
			name:     "cancel when process did not succeed",
			waitErr:  errors.New("signal: killed"),
			ctxErr:   context.Canceled,
			wantCode: 1,
			wantErr:  context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			code, err := outcomeAfterStop(tt.waitErr, tt.ctxErr)
			if code != tt.wantCode {
				t.Fatalf("code = %d, want %d", code, tt.wantCode)
			}
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want errors.Is(..., %v)", err, tt.wantErr)
			}
		})
	}
}
