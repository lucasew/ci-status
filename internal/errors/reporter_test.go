package errors

import (
	"bytes"
	"errors"
	"testing"
)

func TestReport(t *testing.T) {
	buf := new(bytes.Buffer)
	old := Writer
	Writer = buf
	defer func() { Writer = old }()

	err := errors.New("test error")
	Report(err)

	got := buf.String()
	want := "Error: test error\n"
	if got != want {
		t.Errorf("Report() = %q, want %q", got, want)
	}
}

func TestWarn(t *testing.T) {
	buf := new(bytes.Buffer)
	old := Writer
	Writer = buf
	defer func() { Writer = old }()

	err := errors.New("test warning")
	Warn(err)

	got := buf.String()
	want := "Warning: test warning\n"
	if got != want {
		t.Errorf("Warn() = %q, want %q", got, want)
	}
}

func TestWarnf(t *testing.T) {
	buf := new(bytes.Buffer)
	old := Writer
	Writer = buf
	defer func() { Writer = old }()

	Warnf("foo %s", "bar")

	got := buf.String()
	want := "Warning: foo bar\n"
	if got != want {
		t.Errorf("Warnf() = %q, want %q", got, want)
	}
}
