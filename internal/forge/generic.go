package forge

import (
	"fmt"
	"net/url"
	"os"
	"path"
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
	// Trim .git suffix and handle scp-like URLs by replacing the first ':' with '/'
	// This is a common pattern for git remotes, e.g., git@gitea.com:user/repo
	sanitizedURL := strings.TrimSuffix(remoteURL, ".git")
	if !strings.Contains(sanitizedURL, "://") {
		if parts := strings.SplitN(sanitizedURL, ":", 2); len(parts) == 2 {
			sanitizedURL = parts[0] + "/" + parts[1]
		}
	}

	// Use the standard library to parse the URL, which is safer than manual splitting.
	parsedURL, err := url.Parse(sanitizedURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse generic remote URL: %w", err)
	}

	// Clean the path to resolve any directory traversal attempts (e.g., ../)
	cleanedPath := path.Clean(parsedURL.Path)
	parts := strings.Split(strings.Trim(cleanedPath, "/"), "/")

	// We expect at least two parts for owner and repo
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid URL path for remote: %s", remoteURL)
	}

	// The owner and repo are the last two parts of the path
	owner = parts[len(parts)-2]
	repo = parts[len(parts)-1]

	// Final validation to ensure owner and repo are not empty or "."
	if owner == "" || repo == "" || owner == "." || repo == "." {
		return "", "", fmt.Errorf("extracted owner or repo is invalid from: %s", remoteURL)
	}

	return owner, repo, nil
}
