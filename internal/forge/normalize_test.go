package forge_test

import (
	"testing"

	"ci-status/internal/forge"
)

// TestNormalizeRemoteURL_HostCaseAndDefaultPort ensures remote host casing and
// default ports do not break forge detection. ParseGitHubRemote compares hosts
// with == / literal prefixes; without normalization, GitHub.com or :443 remotes
// fail open as "unsupported forge" even when GITHUB_TOKEN is set.
func TestNormalizeRemoteURL_HostCaseAndDefaultPort(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		// which loader should claim the remote
		wantGitHub  bool
		wantGeneric bool
	}{
		{
			name:       "github https default port 443",
			url:        "https://github.com:443/owner/repo.git",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantGitHub: true,
		},
		{
			name:       "github https mixed-case host",
			url:        "https://GitHub.com/owner/repo.git",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantGitHub: true,
		},
		{
			name:       "github scp mixed-case host",
			url:        "git@GitHub.com:owner/repo.git",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantGitHub: true,
		},
		{
			name:       "github ssh URL mixed-case host",
			url:        "ssh://git@GitHub.com/owner/repo.git",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantGitHub: true,
		},
		{
			name:        "gitea keeps non-default port",
			url:         "https://gitea.example.com:3000/owner/repo.git",
			wantOwner:   "owner",
			wantRepo:    "repo",
			wantGeneric: true,
		},
		{
			name:        "gitea mixed-case host still generic",
			url:         "https://Gitea.Example.com/owner/repo.git",
			wantOwner:   "owner",
			wantRepo:    "repo",
			wantGeneric: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := forge.ParseGitHubRemote(tt.url)
			if tt.wantGitHub {
				if err != nil {
					t.Fatalf("ParseGitHubRemote(%q) error: %v", tt.url, err)
				}
				if owner != tt.wantOwner || repo != tt.wantRepo {
					t.Fatalf("ParseGitHubRemote(%q) = (%q,%q), want (%q,%q)",
						tt.url, owner, repo, tt.wantOwner, tt.wantRepo)
				}
			}

			if tt.wantGeneric {
				owner, repo, err = forge.ParseGenericRemote(tt.url)
				if err != nil {
					t.Fatalf("ParseGenericRemote(%q) error: %v", tt.url, err)
				}
				if owner != tt.wantOwner || repo != tt.wantRepo {
					t.Fatalf("ParseGenericRemote(%q) = (%q,%q), want (%q,%q)",
						tt.url, owner, repo, tt.wantOwner, tt.wantRepo)
				}
			}

			gh := forge.LoadGitHub(tt.url)
			gen := forge.LoadGeneric(tt.url)
			if tt.wantGitHub && gh == nil {
				t.Fatalf("LoadGitHub(%q) = nil, want client", tt.url)
			}
			if !tt.wantGitHub && gh != nil {
				t.Fatalf("LoadGitHub(%q) = %#v, want nil", tt.url, gh)
			}
			if tt.wantGeneric && gen == nil {
				t.Fatalf("LoadGeneric(%q) = nil, want client", tt.url)
			}
			if !tt.wantGeneric && gen != nil {
				t.Fatalf("LoadGeneric(%q) = %#v, want nil", tt.url, gen)
			}

			if tt.wantGeneric {
				client, ok := gen.(*forge.GitHubClient)
				if !ok {
					t.Fatalf("LoadGeneric type %T, want *GitHubClient", gen)
				}
				// Non-default port must remain in the API base URL.
				if tt.url == "https://gitea.example.com:3000/owner/repo.git" &&
					client.BaseURL != "https://gitea.example.com:3000/api/v1" {
					t.Fatalf("BaseURL = %q, want port preserved", client.BaseURL)
				}
				if tt.url == "https://Gitea.Example.com/owner/repo.git" &&
					client.BaseURL != "https://gitea.example.com/api/v1" {
					t.Fatalf("BaseURL = %q, want lowercased host", client.BaseURL)
				}
			}
		})
	}
}
