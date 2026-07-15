package forge

import (
	"strings"
	"testing"
)

func TestDetectClientFromURL_MissingToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	cases := []struct {
		name    string
		url     string
		wantSub string
	}{
		{
			name:    "github https",
			url:     "https://github.com/owner/repo.git",
			wantSub: "GITHUB_TOKEN not set",
		},
		{
			name:    "github ssh",
			url:     "git@github.com:owner/repo.git",
			wantSub: "GITHUB_TOKEN not set",
		},
		{
			name:    "gitea https",
			url:     "https://gitea.example.com/owner/repo.git",
			wantSub: "GITHUB_TOKEN not set",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			client, err := detectClientFromURL(tt.url, "")
			if client != nil {
				t.Fatalf("expected nil client without token, got %#v", client)
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("error %q should contain %q", err.Error(), tt.wantSub)
			}
			if strings.Contains(err.Error(), "no supported forge") {
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
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no supported forge") {
		t.Fatalf("want unsupported forge message, got %v", err)
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
