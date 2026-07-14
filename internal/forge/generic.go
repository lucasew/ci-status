package forge

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// LoadGeneric is a strategy to initialize a ForgeClient for generic Gitea/Forgejo instances.
// It assumes the forge supports a GitHub-compatible API at `/api/v1`.
// It explicitly rejects GitHub URLs to prevent fallback loops or incorrect client initialization.
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

// getHostAndScheme extracts the host and scheme from the remote URL.
// For SSH URLs, it assumes "https" as the API scheme.
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

// ParseGenericRemote extracts the owner and repository from a generic remote URL.
// HTTP(S) and ssh:// URLs use the path only (host is never treated as owner).
// SCP-like syntax (git@host:path) is normalized by replacing the first colon with a slash.
// Owner and repo are always the last two path segments (supports nested groups).
func ParseGenericRemote(remoteURL string) (owner, repo string, err error) {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	remoteURL = strings.TrimSuffix(remoteURL, "/")

	var pathParts []string

	if strings.Contains(remoteURL, "://") {
		// http(s):// and ssh:// — parse with net/url so host/scheme stay out of the path.
		u, parseErr := url.Parse(remoteURL)
		if parseErr != nil {
			return "", "", fmt.Errorf("cannot parse generic remote: %s", remoteURL)
		}
		pathParts = strings.FieldsFunc(u.Path, func(r rune) bool { return r == '/' })
	} else {
		// SCP-like: git@host:owner/repo → treat host:path separator as '/'.
		cleanURL := strings.Replace(remoteURL, ":", "/", 1)
		pathParts = strings.FieldsFunc(cleanURL, func(r rune) bool { return r == '/' })
		// Drop user@host style first segment when present (e.g. "git@host").
		if len(pathParts) > 0 && strings.Contains(pathParts[0], "@") {
			pathParts = pathParts[1:]
		}
	}

	if len(pathParts) < 2 {
		return "", "", fmt.Errorf("cannot parse generic remote: %s", remoteURL)
	}

	return pathParts[len(pathParts)-2], pathParts[len(pathParts)-1], nil
}
