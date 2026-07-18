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

func TestLoadGenericBaseURL(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	tests := []struct {
		name    string
		url     string
		wantURL string
		wantNil bool
	}{
		{
			name:    "https keeps non-default port",
			url:     "https://gitea.example.com:3000/owner/repo.git",
			wantURL: "https://gitea.example.com:3000/api/v1",
		},
		{
			name:    "https default port omitted in remote stays host only",
			url:     "https://gitea.example.com/owner/repo.git",
			wantURL: "https://gitea.example.com/api/v1",
		},
		{
			name:    "ssh:// without port uses https host",
			url:     "ssh://git@gitea.example.com/owner/repo.git",
			wantURL: "https://gitea.example.com/api/v1",
		},
		{
			name:    "ssh:// with SSH port does not reuse port for API",
			url:     "ssh://git@gitea.example.com:2222/owner/repo.git",
			wantURL: "https://gitea.example.com/api/v1",
		},
		{
			name:    "scp-like ssh",
			url:     "git@gitea.example.com:owner/repo.git",
			wantURL: "https://gitea.example.com/api/v1",
		},
		{
			name:    "rejects github https",
			url:     "https://github.com/owner/repo.git",
			wantNil: true,
		},
		{
			name:    "rejects github https with port",
			url:     "https://github.com:443/owner/repo.git",
			wantNil: true,
		},
		{
			name:    "rejects github scp",
			url:     "git@github.com:owner/repo.git",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := forge.LoadGeneric(tt.url)
			if tt.wantNil {
				if client != nil {
					t.Fatalf("LoadGeneric(%q) = %#v, want nil", tt.url, client)
				}
				return
			}
			if client == nil {
				t.Fatalf("LoadGeneric(%q) = nil, want client", tt.url)
			}
			gh, ok := client.(*forge.GitHubClient)
			if !ok {
				t.Fatalf("LoadGeneric(%q) type %T, want *forge.GitHubClient", tt.url, client)
			}
			if gh.BaseURL != tt.wantURL {
				t.Fatalf("BaseURL = %q, want %q", gh.BaseURL, tt.wantURL)
			}
			if gh.Owner != "owner" || gh.Repo != "repo" {
				t.Fatalf("owner/repo = %s/%s, want owner/repo", gh.Owner, gh.Repo)
			}
		})
	}
}

func TestLoadGenericRequiresToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	if client := forge.LoadGeneric("https://gitea.example.com/owner/repo.git"); client != nil {
		t.Fatalf("expected nil without token, got %#v", client)
	}
}
