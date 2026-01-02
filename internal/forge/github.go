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

type GitHubClient struct {
	Token   string
	Owner   string
	Repo    string
	BaseURL string
}

func NewGitHubClient(token, owner, repo string) *GitHubClient {
	return &GitHubClient{
		Token: token,
		Owner: owner,
		Repo:  repo,
	}
}

func (c *GitHubClient) SetStatus(ctx context.Context, opts StatusOpts) error {
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/repos/%s/%s/statuses/%s", baseURL, c.Owner, c.Repo, opts.Commit)

	state := string(opts.State)
	// GitHub API specifically treats "running" as "pending" with a description,
	// but only if we are talking to actual GitHub. For compatible forges (Gitea),
	// they might support "running" or we might want to stick to "pending".
	// The existing logic checks for prefix https://api.github.com.
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

	// Sanitize token to prevent header injection.
	sanitizedToken := strings.NewReplacer("\n", "", "\r", "").Replace(c.Token)
	if sanitizedToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sanitizedToken))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

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

func LoadGitHub(url string) ForgeClient {
	owner, repo, err := ParseGitHubRemote(url)
	if err != nil {
		return nil
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// Log warning? For now just return nil or let it fail later?
		// The requirement is "nil se não encaixa".
		// If it parses as GitHub, it "fits". But if we have no token, we can't use it.
		// However, returning nil might trigger fallback.
		// If it IS github.com, we shouldn't fallback to generic.
		// But if no token, we can't do anything.
		// Let's return nil if no token, effectively disabling it.
		// Or we return a client that will fail later.
		// Based on `run.go` logic: if token missing -> noop.
		// So we should probably return nil here if we want noop.
		return nil
	}
	return NewGitHubClient(token, owner, repo)
}

func ParseGitHubRemote(remoteURL string) (owner, repo string, err error) {
	// Supports:
	// https://github.com/owner/repo.git
	// https://github.com/owner/repo
	// git@github.com:owner/repo.git
	// https://x-access-token:...@github.com/owner/repo.git

	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	remoteURL = strings.TrimSuffix(remoteURL, "/")

	// Try parsing as URL first to handle auth and other schemes robustly
	if u, err := url.Parse(remoteURL); err == nil {
		if u.Host == "github.com" {
			path := strings.TrimPrefix(u.Path, "/")
			parts := strings.Split(path, "/")
			if len(parts) == 2 {
				return parts[0], parts[1], nil
			}
		}
	}

	// Fallback/Legacy handling for formats url.Parse might misinterpret (like git@github.com:...)
	// although url.Parse usually fails or puts it in path for SCP-like syntax.

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
