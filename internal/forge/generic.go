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

	return parts[len(parts)-2], parts[len(parts)-1], nil
}
