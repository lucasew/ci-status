package forge

import (
	"errors"
	"testing"
)

func TestDetectClientFromURL_MissingToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	cases := []struct {
		name string
		url  string
	}{
		{
			name: "github https",
			url:  "https://github.com/owner/repo.git",
		},
		{
			name: "github ssh",
			url:  "git@github.com:owner/repo.git",
		},
		{
			name: "gitea https",
			url:  "https://gitea.example.com/owner/repo.git",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			client, err := detectClientFromURL(tt.url, "")
			if client != nil {
				t.Fatalf("expected nil client without token, got %#v", client)
			}
			if !errors.Is(err, ErrGitHubTokenNotSet) {
				t.Fatalf("err = %v, want ErrGitHubTokenNotSet", err)
			}
			if errors.Is(err, ErrNoSupportedForge) {
				t.Fatalf("error should not claim unsupported forge: %v", err)
			}
		})
	}
}

func TestDetectClientFromURL_UnsupportedRemote(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	client, err := detectClientFromURL("not-a-remote", "")
	if client != nil {
		t.Fatalf("expected nil client, got %#v", client)
	}
	if !errors.Is(err, ErrNoSupportedForge) {
		t.Fatalf("err = %v, want ErrNoSupportedForge", err)
	}
}

func TestDetectClientFromURL_WithToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	client, err := detectClientFromURL("https://github.com/owner/repo.git", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client with token and github remote")
	}

	generic, err := detectClientFromURL("https://gitea.example.com/owner/repo.git", "")
	if err != nil {
		t.Fatalf("unexpected error for generic: %v", err)
	}
	if generic == nil {
		t.Fatal("expected generic client with token")
	}
}

func TestDetectClientFromURL_OverrideGitHub(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	client, err := detectClientFromURL("https://github.com/owner/repo.git", "github")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client with override=github")
	}
}

func TestDetectClientFromURL_OverrideGitHubNoFallthrough(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	// Generic/Gitea remotes must not be used when the user forced --forge github.
	client, err := detectClientFromURL("https://gitea.example.com/owner/repo.git", "github")
	if client != nil {
		t.Fatalf("expected nil client for non-GitHub remote with github override, got %#v", client)
	}
	if !errors.Is(err, ErrCouldNotLoadGitHubClient) {
		t.Fatalf("err = %v, want ErrCouldNotLoadGitHubClient", err)
	}
}

func TestDetectClientFromURL_UnsupportedOverride(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")

	client, err := detectClientFromURL("https://github.com/owner/repo.git", "gitlab")
	if client != nil {
		t.Fatalf("expected nil client for unsupported override, got %#v", client)
	}
	if !errors.Is(err, ErrUnsupportedForgeOverride) {
		t.Fatalf("err = %v, want ErrUnsupportedForgeOverride", err)
	}
}

func TestDetectClientFromURL_OverrideGitHubMissingToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	client, err := detectClientFromURL("https://github.com/owner/repo.git", "github")
	if client != nil {
		t.Fatalf("expected nil client without token, got %#v", client)
	}
	if !errors.Is(err, ErrGitHubTokenNotSet) {
		t.Fatalf("err = %v, want ErrGitHubTokenNotSet", err)
	}
}
