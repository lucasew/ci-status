package main

import (
	"strings"
	"testing"

	"ci-status/internal/forge"
)

func TestParseState(t *testing.T) {
	valid := []string{
		string(forge.StatePending),
		string(forge.StateSuccess),
		string(forge.StateFailure),
		string(forge.StateError),
		string(forge.StateRunning),
	}
	for _, s := range valid {
		got, err := parseState(s)
		if err != nil {
			t.Fatalf("parseState(%q) unexpected error: %v", s, err)
		}
		if got != forge.State(s) {
			t.Fatalf("parseState(%q) = %q, want %q", s, got, s)
		}
	}

	_, err := parseState("bogon")
	if err == nil {
		t.Fatal("parseState(bogon) expected error")
	}
	if !strings.Contains(err.Error(), "invalid state") {
		t.Fatalf("error %q should mention invalid state", err.Error())
	}
}

func TestExecuteSet_NotCI_Noop(t *testing.T) {
	t.Setenv("CI", "")
	err := executeSet(SetConfig{
		ContextName: "lint",
		State:       "success",
		Silent:      true,
	})
	if err != nil {
		t.Fatalf("outside CI, set should noop successfully, got %v", err)
	}
}

func TestExecuteSet_CI_MissingToken(t *testing.T) {
	t.Setenv("CI", "true")
	t.Setenv("GITHUB_TOKEN", "")
	// Force a github-ish remote via git is hard; DetectClient uses real git remote
	// of this checkout (github.com/lucasew/ci-status), so missing token must error.
	err := executeSet(SetConfig{
		ContextName: "lint",
		State:       "success",
		Commit:      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Silent:      true,
	})
	if err == nil {
		t.Fatal("expected error when CI is set but credentials/client unavailable")
	}
}
