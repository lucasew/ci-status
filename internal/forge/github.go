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

// SetStatus updates the commit status on GitHub.
// Note: It automatically maps the 'running' state to 'pending' because GitHub's API
// does not natively support a 'running' state for commit statuses (only Checks API does).
func (c *GitHubClient) SetStatus(ctx context.Context, opts StatusOpts) error {
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/repos/%s/%s/statuses/%s", baseURL, c.Owner, c.Repo, opts.Commit)

	state := string(opts.State)
	// GitHub API treats "running" as "pending". We map it here unless we are targeting
	// a different forge that might support it (checked via BaseURL).
	// If BaseURL is default (empty) or specifically GitHub API, we map to pending.
	if opts.State == StateRunning && (strings.HasPrefix(url, "https://api.github.com") || c.BaseURL == "") {
		state = string(StatePending)
	}

	body := map[string]string{
		"state":       state,
		"description": opts.Description,
		"context":     opts.Context,
	}
	if opts.TargetURL != "" {
		body["target_url"] = opts.TargetURL
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
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
			path := strings.TrimPrefix(u.Path, "/")
			parts := strings.Split(path, "/")
			if len(parts) == 2 {
				return parts[0], parts[1], nil
			}
		}
	}

	// Fallback handling for formats that url.Parse might misinterpret (e.g., SCP-like syntax).
	if strings.HasPrefix(remoteURL, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "https://github.com/"), "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid https github url: %s", remoteURL)
		}
		return parts[0], parts[1], nil
	}

	if strings.HasPrefix(remoteURL, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "git@github.com:"), "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid ssh github url: %s", remoteURL)
		}
		return parts[0], parts[1], nil
	}

	if strings.HasPrefix(remoteURL, "ssh://git@github.com/") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "ssh://git@github.com/"), "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid ssh github url: %s", remoteURL)
		}
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unrecognized github url format: %s", remoteURL)
}
