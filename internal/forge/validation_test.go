package forge

import (
	"strings"
	"testing"
)

func TestPathTraversal_Generic(t *testing.T) {
	maliciousURL := "https://example.com/owner/.."

	_, _, err := ParseGenericRemote(maliciousURL)
	if err == nil {
		t.Fatal("Expected error due to path traversal, but got nil")
	}
	if !strings.Contains(err.Error(), "invalid repo in generic remote") {
		t.Errorf("Expected error message to mention invalid repo, got: %v", err)
	}
}

func TestQueryInjection_Generic(t *testing.T) {
	maliciousURL := "https://example.com/owner/repo?x=1"

	_, _, err := ParseGenericRemote(maliciousURL)
	if err == nil {
		t.Fatal("Expected error due to query injection, but got nil")
	}
}

func TestPathTraversal_GitHub(t *testing.T) {
	maliciousURL := "https://github.com/owner/.."

	_, _, err := ParseGitHubRemote(maliciousURL)
	if err == nil {
		t.Fatal("Expected error due to path traversal, but got nil")
	}
}
