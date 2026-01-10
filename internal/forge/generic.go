package forge

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func LoadGeneric(remoteURL string) ForgeClient {
	owner, repo, err := ParseGenericRemote(remoteURL)
	if err != nil {
		return nil
	}

	// We need to determine the BaseURL.
	// Assume Gitea/Forgejo compatible API at /api/v1
	host, scheme := getHostAndScheme(remoteURL)
	if host == "" {
		return nil
	}

	// Prevent generic loader from taking over GitHub URLs if the specific loader failed
	if host == "github.com" || host == "api.github.com" {
		return nil
	}

	// Use GITHUB_TOKEN as fallback for generic forge token
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil
	}

	client := NewGitHubClient(token, owner, repo)
	client.BaseURL = fmt.Sprintf("%s://%s/api/v1", scheme, host)

	return client
}

func getHostAndScheme(remoteURL string) (string, string) {
	// Handle HTTP/HTTPS
	if strings.HasPrefix(remoteURL, "http://") || strings.HasPrefix(remoteURL, "https://") {
		u, err := url.Parse(remoteURL)
		if err != nil {
			return "", ""
		}
		return u.Host, u.Scheme
	}

	// Handle SSH (git@host:...) or (ssh://...)
	if strings.HasPrefix(remoteURL, "ssh://") {
		u, err := url.Parse(remoteURL)
		if err != nil {
			return "", ""
		}
		return u.Host, "https" // Assume HTTPS for API
	}

	// Scp-like: git@host:path
	if strings.Contains(remoteURL, "@") && strings.Contains(remoteURL, ":") {
		parts := strings.Split(remoteURL, "@")
		if len(parts) > 1 {
			domainPath := parts[1]
			domainParts := strings.Split(domainPath, ":")
			if len(domainParts) > 0 {
				return domainParts[0], "https" // Assume HTTPS for API
			}
		}
	}

	return "", ""
}

func ParseGenericRemote(remoteURL string) (owner, repo string, err error) {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// SECURITY: The previous implementation used string splitting on the entire
	// URL, which could lead to path injection if a malformed URL was provided
	// (e.g., `https://example.com` would be parsed as owner=`https:`, repo=`example.com`).
	//
	// To fix this, we now use `net/url.Parse` for robust parsing. We also
	// handle scp-like git URLs by converting them to a ssh:// scheme first.
	//
	// Example of scp-like URL: git@gitea.com:user/repo
	// Converted to: ssh://git@gitea.com/user/repo
	if !strings.Contains(remoteURL, "://") && strings.Contains(remoteURL, ":") {
		remoteURL = "ssh://" + strings.Replace(remoteURL, ":", "/", 1)
	}

	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse remote URL: %w", err)
	}

	// We expect the path to be in the format `/<owner>/<repo>`
	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid path in remote URL, expected owner/repo: %s", u.Path)
	}

	// Take the last two parts of the path as owner and repo
	return parts[len(parts)-2], parts[len(parts)-1], nil
}
