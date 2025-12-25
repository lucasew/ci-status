package forge_test

import (
	"testing"

	"ci-status/internal/forge"
)

func TestParseGitHubRemote(t *testing.T) {
	tests := []struct {
		url   string
		owner string
		repo  string
		err   bool
	}{
		{"https://github.com/owner/repo.git", "owner", "repo", false},
		{"https://github.com/owner/repo", "owner", "repo", false},
		{"git@github.com:owner/repo.git", "owner", "repo", false},
		{"git@github.com:owner/repo", "owner", "repo", false},
		{"ssh://git@github.com/owner/repo.git", "owner", "repo", false},
		{"https://gitlab.com/owner/repo.git", "", "", true},
		{"invalid", "", "", true},
	}

	for _, tt := range tests {
		owner, repo, err := forge.ParseGitHubRemote(tt.url)
		if tt.err {
			if err == nil {
				t.Errorf("ParseGitHubRemote(%s) expected error, got nil", tt.url)
			}
		} else {
			if err != nil {
				t.Errorf("ParseGitHubRemote(%s) expected no error, got %v", tt.url, err)
			}
			if owner != tt.owner {
				t.Errorf("ParseGitHubRemote(%s) expected owner %s, got %s", tt.url, tt.owner, owner)
			}
			if repo != tt.repo {
				t.Errorf("ParseGitHubRemote(%s) expected repo %s, got %s", tt.url, tt.repo, repo)
			}
		}
	}
}
