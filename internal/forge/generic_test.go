package forge

import (
	"testing"
)

func TestParseGenericRemote_PathTraversal(t *testing.T) {
	tests := []struct {
		url           string
		expectedOwner string
		expectedRepo  string
		description   string
		shouldError   bool
	}{
		{
			url:           "https://example.com/owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			description:   "Normal URL",
		},
		{
			url:           "https://example.com/owner/repo/../../attacker/project",
			expectedOwner: "attacker",
			expectedRepo:  "project",
			description:   "Path traversal resolved (attacker/project)",
		},
		{
			url:           "https://example.com/owner/repo/../sub",
			expectedOwner: "owner", // /owner/repo/../sub -> /owner/sub
			expectedRepo:  "sub",
			description:   "Path traversal resolved",
		},
		{
			url:           "https://example.com/../repo",
			shouldError:   true,
			description:   "Path traversal to root with insufficient parts",
		},
		{
			url: "git@github.com:owner/repo",
			expectedOwner: "owner",
			expectedRepo: "repo",
			description: "SCP-style URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			owner, repo, err := ParseGenericRemote(tt.url)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if owner != tt.expectedOwner {
				t.Errorf("expected owner %q, got %q", tt.expectedOwner, owner)
			}
			if repo != tt.expectedRepo {
				t.Errorf("expected repo %q, got %q", tt.expectedRepo, repo)
			}
		})
	}
}
