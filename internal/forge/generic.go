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
// It normalizes SCP-like syntax (git@host:path) by replacing the first colon with a slash,
// allowing consistent path splitting.
// Security Note: This implementation aims to be robust against path traversal attempts
// by strictly using the last two path segments as owner and repo.
func ParseGenericRemote(remoteURL string) (owner, repo string, err error) {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	cleanURL := remoteURL
	// Handle scp-like syntax which uses ':' as separator
	// e.g. git@github.com:owner/repo -> git@github.com/owner/repo
	// We want to treat ':' as '/' if it's not part of the protocol scheme (://)
	if idx := strings.Index(cleanURL, "://"); idx == -1 {
		// No protocol scheme, assume ssh/scp-like
		// Replace the FIRST ':' which separates host from path (or port?)
		// git@host:owner/repo
		cleanURL = strings.Replace(cleanURL, ":", "/", 1)
	}

	parts := strings.FieldsFunc(cleanURL, func(r rune) bool { return r == '/' })
	if len(parts) < 2 {
		return "", "", fmt.Errorf("cannot parse generic remote: %s", remoteURL)
	}

	// Taking the last two parts handles cases like ssh://host/group/subgroup/owner/repo
	owner = parts[len(parts)-2]
	repo = parts[len(parts)-1]

	if err := validateRepoName(owner); err != nil {
		return "", "", fmt.Errorf("invalid owner in generic remote: %w", err)
	}
	if err := validateRepoName(repo); err != nil {
		return "", "", fmt.Errorf("invalid repo in generic remote: %w", err)
	}

	return owner, repo, nil
}
