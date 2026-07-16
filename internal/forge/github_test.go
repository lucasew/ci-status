package forge_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestSetStatusMapsRunningToPending ensures StateRunning is never sent to a
// GitHub-compatible statuses API (including custom BaseURL hosts like Gitea).
func TestSetStatusMapsRunningToPending(t *testing.T) {
	var gotState string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var payload map[string]string
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("unmarshal: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		gotState = payload["state"]
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)

	client := forge.NewGitHubClient("token", "owner", "repo")
	client.BaseURL = srv.URL

	err := client.SetStatus(context.Background(), forge.StatusOpts{
		Commit:      "abc123",
		Context:     "lint",
		State:       forge.StateRunning,
		Description: "Running...",
	})
	if err != nil {
		t.Fatalf("SetStatus: %v", err)
	}
	if gotState != string(forge.StatePending) {
		t.Fatalf("expected state %q, got %q", forge.StatePending, gotState)
	}
}

// TestSetStatusTruncatesLongFields ensures description/context never exceed
// GitHub's 140/100 character limits (API returns 422 otherwise).
func TestSetStatusTruncatesLongFields(t *testing.T) {
	var got map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("unmarshal: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)

	longDesc := strings.Repeat("d", 200)
	longCtx := strings.Repeat("c", 150)

	client := forge.NewGitHubClient("token", "owner", "repo")
	client.BaseURL = srv.URL

	err := client.SetStatus(context.Background(), forge.StatusOpts{
		Commit:      "abc123",
		Context:     longCtx,
		State:       forge.StateSuccess,
		Description: longDesc,
	})
	if err != nil {
		t.Fatalf("SetStatus: %v", err)
	}
	if got == nil {
		t.Fatal("expected request body")
	}
	if n := len([]rune(got["description"])); n != 140 {
		t.Fatalf("description rune length = %d, want 140 (got %q)", n, got["description"])
	}
	if !strings.HasSuffix(got["description"], "…") {
		t.Fatalf("description should end with ellipsis, got %q", got["description"])
	}
	if n := len([]rune(got["context"])); n != 100 {
		t.Fatalf("context rune length = %d, want 100 (got %q)", n, got["context"])
	}
	if !strings.HasSuffix(got["context"], "…") {
		t.Fatalf("context should end with ellipsis, got %q", got["context"])
	}
}

func TestSetStatusKeepsShortFields(t *testing.T) {
	var got map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Errorf("unmarshal: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)

	client := forge.NewGitHubClient("token", "owner", "repo")
	client.BaseURL = srv.URL

	err := client.SetStatus(context.Background(), forge.StatusOpts{
		Commit:      "abc123",
		Context:     "ci/lint",
		State:       forge.StateSuccess,
		Description: "Passed",
	})
	if err != nil {
		t.Fatalf("SetStatus: %v", err)
	}
	if got["description"] != "Passed" || got["context"] != "ci/lint" {
		t.Fatalf("short fields changed: description=%q context=%q", got["description"], got["context"])
	}
}

func asGitHubClient(t *testing.T, c forge.ForgeClient) *forge.GitHubClient {
	t.Helper()
	gh, ok := c.(*forge.GitHubClient)
	if !ok || gh == nil {
		t.Fatalf("expected *GitHubClient, got %#v", c)
	}
	return gh
}

func TestLoadGitHub_GitHubDotComRemote(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "tok")
	t.Setenv("GITHUB_API_URL", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITHUB_REPOSITORY", "")

	c := forge.LoadGitHub("https://github.com/acme/app.git")
	gh := asGitHubClient(t, c)
	if gh.Owner != "acme" || gh.Repo != "app" {
		t.Fatalf("owner/repo = %s/%s, want acme/app", gh.Owner, gh.Repo)
	}
	if gh.BaseURL != "" {
		t.Fatalf("BaseURL = %q, want empty (default api.github.com)", gh.BaseURL)
	}
}

func TestLoadGitHub_UsesAPIURLAndRepoEnvOnGHES(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "tok")
	t.Setenv("GITHUB_API_URL", "https://ghe.example.com/api/v3/")
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_REPOSITORY", "acme/app")

	// Non-github.com remote: without Actions env this would fall through to generic.
	c := forge.LoadGitHub("https://ghe.example.com/acme/app.git")
	gh := asGitHubClient(t, c)
	if gh.Owner != "acme" || gh.Repo != "app" {
		t.Fatalf("owner/repo = %s/%s, want acme/app from GITHUB_REPOSITORY", gh.Owner, gh.Repo)
	}
	if gh.BaseURL != "https://ghe.example.com/api/v3" {
		t.Fatalf("BaseURL = %q, want GHES api/v3 without trailing slash", gh.BaseURL)
	}
}

func TestLoadGitHub_APIURLOnGitHubDotComRemote(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "tok")
	t.Setenv("GITHUB_API_URL", "https://api.github.com")
	t.Setenv("GITHUB_ACTIONS", "true")
	// Env repo must not override a parsed github.com remote.
	t.Setenv("GITHUB_REPOSITORY", "other/other")

	c := forge.LoadGitHub("https://github.com/acme/app.git")
	gh := asGitHubClient(t, c)
	if gh.Owner != "acme" || gh.Repo != "app" {
		t.Fatalf("owner/repo = %s/%s, want acme/app from remote", gh.Owner, gh.Repo)
	}
	if gh.BaseURL != "https://api.github.com" {
		t.Fatalf("BaseURL = %q, want GITHUB_API_URL", gh.BaseURL)
	}
}

func TestLoadGitHub_DoesNotClaimGiteaWithoutActionsEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "tok")
	t.Setenv("GITHUB_API_URL", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITHUB_REPOSITORY", "acme/app")

	if c := forge.LoadGitHub("https://gitea.example.com/acme/app.git"); c != nil {
		t.Fatalf("expected nil so LoadGeneric can handle Gitea, got %#v", c)
	}
}

func TestLoadGitHub_MissingToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GITHUB_API_URL", "https://api.github.com")
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_REPOSITORY", "acme/app")

	if c := forge.LoadGitHub("https://github.com/acme/app.git"); c != nil {
		t.Fatalf("expected nil without token, got %#v", c)
	}
}
