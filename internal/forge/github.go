package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type GitHubClient struct {
	Token string
	Owner string
	Repo  string
}

func NewGitHubClient(token, owner, repo string) *GitHubClient {
	return &GitHubClient{
		Token: token,
		Owner: owner,
		Repo:  repo,
	}
}

func (c *GitHubClient) SetStatus(ctx context.Context, opts StatusOpts) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/statuses/%s", c.Owner, c.Repo, opts.Commit)

	if opts.State == StateRunning && strings.HasPrefix(url, "https://api.github.com") {
		fmt.Fprintln(os.Stderr, "Warning: Sending 'running' state to api.github.com is not supported and will likely fail.")
	}

	body := map[string]string{
		"state":       string(opts.State),
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

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github api error: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

func ParseGitHubRemote(remoteURL string) (owner, repo string, err error) {
	// Supports:
	// https://github.com/owner/repo.git
	// https://github.com/owner/repo
	// git@github.com:owner/repo.git

	remoteURL = strings.TrimSuffix(remoteURL, ".git")

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

	return "", "", fmt.Errorf("unrecognized github url format: %s", remoteURL)
}
