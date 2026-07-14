package forge_test

import (
	"testing"

	"ci-status/internal/forge"
)

func TestParseGenericRemote(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		owner string
		repo  string
		err   bool
	}{
		{
			name:  "https gitea",
			url:   "https://gitea.example.com/owner/repo.git",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "https without .git",
			url:   "https://git.example.com/org/project",
			owner: "org",
			repo:  "project",
		},
		{
			name:  "scp-like ssh",
			url:   "git@gitea.example.com:owner/repo.git",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "ssh url",
			url:   "ssh://git@git.example.com/owner/repo.git",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "nested group path uses last two segments",
			url:   "https://git.example.com/group/subgroup/owner/repo",
			owner: "owner",
			repo:  "repo",
		},
		{
			name: "too few path segments",
			url:  "https://git.example.com/onlyrepo",
			err:  true,
		},
		{
			name: "empty",
			url:  "",
			err:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := forge.ParseGenericRemote(tt.url)
			if tt.err {
				if err == nil {
					t.Fatalf("ParseGenericRemote(%q) expected error, got owner=%q repo=%q", tt.url, owner, repo)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseGenericRemote(%q) unexpected error: %v", tt.url, err)
			}
			if owner != tt.owner || repo != tt.repo {
				t.Fatalf("ParseGenericRemote(%q) = (%q, %q), want (%q, %q)", tt.url, owner, repo, tt.owner, tt.repo)
			}
		})
	}
}
