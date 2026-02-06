package errors

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestReport(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(buf, nil))
	slog.SetDefault(logger)

	err := errors.New("test error")
	Report(err)

	got := buf.String()
	// slog TextHandler format: "time=... level=ERROR msg=\"error reported\" err=\"test error\"\n"
	if !strings.Contains(got, "level=ERROR") || !strings.Contains(got, "msg=\"error reported\"") || !strings.Contains(got, "err=\"test error\"") {
		t.Errorf("Report() output unexpected: %s", got)
	}
}

func TestWarn(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(buf, nil))
	slog.SetDefault(logger)

	err := errors.New("test warning")
	Warn(err)

	got := buf.String()
	if !strings.Contains(got, "level=WARN") || !strings.Contains(got, "msg=\"warning reported\"") || !strings.Contains(got, "err=\"test warning\"") {
		t.Errorf("Warn() output unexpected: %s", got)
	}
}

func TestWarnf(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(buf, nil))
	slog.SetDefault(logger)

	Warnf("foo %s", "bar")

	got := buf.String()
	if !strings.Contains(got, "level=WARN") || !strings.Contains(got, "msg=\"foo bar\"") {
		t.Errorf("Warnf() output unexpected: %s", got)
	}
}
