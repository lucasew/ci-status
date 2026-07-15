package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// GitHubClient implements the ForgeClient interface for GitHub and compatible APIs.
// It handles authentication via personal access tokens and status updates.
type GitHubClient struct {
	Token   string
	Owner   string
	Repo    string
	BaseURL string
}

// NewGitHubClient creates a new instance of GitHubClient.
func NewGitHubClient(token, owner, repo string) *GitHubClient {
	return &GitHubClient{
		Token: token,
		Owner: owner,
		Repo:  repo,
	}
}

// GitHub commit-status field limits (REST: create a commit status).
// Exceeding them makes the API return 422 and the whole status update fail.
const (
	maxStatusDescriptionLen = 140
	maxStatusContextLen     = 100
)

// truncateRunes shortens s to at most max runes. When truncated, the last rune
// is replaced with an ellipsis so callers can see the value was cut.
func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

// SetStatus updates the commit status on GitHub and GitHub-compatible APIs
// (GitHub Enterprise, Gitea/Forgejo via /api/v1). Commit status endpoints only
// accept error|failure|pending|success, so StateRunning is always mapped to pending.
// Description is capped at 140 characters and context at 100 (GitHub API limits).
func (c *GitHubClient) SetStatus(ctx context.Context, opts StatusOpts) error {
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	statusURL := fmt.Sprintf("%s/repos/%s/%s/statuses/%s", baseURL, c.Owner, c.Repo, opts.Commit)

	state := string(opts.State)
	if opts.State == StateRunning {
		state = string(StatePending)
	}

	body := map[string]string{
		"state":       state,
		"description": truncateRunes(opts.Description, maxStatusDescriptionLen),
		"context":     truncateRunes(opts.Context, maxStatusContextLen),
	}
	if opts.TargetURL != "" {
		body["target_url"] = opts.TargetURL
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", statusURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Sanitize token to prevent header injection vulnerabilities.
	sanitizedToken := strings.NewReplacer("\n", "", "\r", "").Replace(c.Token)
	if sanitizedToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sanitizedToken))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Use a custom client with timeout to prevent hanging requests.
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github api error: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

// LoadGitHub is a strategy to initialize a GitHubClient if the URL matches a GitHub repository.
// It requires the GITHUB_TOKEN environment variable to be set.
func LoadGitHub(url string) ForgeClient {
	owner, repo, err := ParseGitHubRemote(url)
	if err != nil {
		return nil
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// If the token is missing, we cannot interact with the API, so we return nil.
		return nil
	}
	return NewGitHubClient(token, owner, repo)
}

// ParseGitHubRemote extracts the owner and repository from a GitHub remote URL.
// It supports standard HTTPS, SSH, and token-embedded HTTPS formats.
// Supported formats:
// - https://github.com/owner/repo.git
// - git@github.com:owner/repo.git
// - ssh://git@github.com/owner/repo.git
func ParseGitHubRemote(remoteURL string) (owner, repo string, err error) {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	remoteURL = strings.TrimSuffix(remoteURL, "/")

	// Try parsing as URL first to handle auth and other schemes robustly.
	if u, err := url.Parse(remoteURL); err == nil {
		if u.Host == "github.com" {
			urlPath := strings.TrimPrefix(u.Path, "/")
			parts := strings.Split(urlPath, "/")
			if len(parts) == 2 {
				return parts[0], parts[1], nil
			}
		}
	}

	// Fallback handling for formats that url.Parse might misinterpret (e.g., SCP-like syntax).
	fallbackFormats := []struct {
		prefix string
		format string
	}{
		{"https://github.com/", "https github url"},
		{"git@github.com:", "ssh github url"},
		{"ssh://git@github.com/", "ssh github url"},
	}

	for _, f := range fallbackFormats {
		if strings.HasPrefix(remoteURL, f.prefix) {
			parts := strings.Split(strings.TrimPrefix(remoteURL, f.prefix), "/")
			if len(parts) != 2 {
				return "", "", fmt.Errorf("invalid %s: %s", f.format, remoteURL)
			}
			return parts[0], parts[1], nil
		}
	}

	return "", "", fmt.Errorf("unrecognized github url format: %s", remoteURL)
}
