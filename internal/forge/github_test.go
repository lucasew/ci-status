package forge_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
