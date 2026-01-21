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
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// Normalize SCP-like syntax (git@host:path -> ssh://git@host/path) to make it url.Parse friendly
	// If it doesn't have ://, it might be scp-style or just a local path?
	// But local paths are usually not remotes we care about here?
	// We should try to prepend ssh:// if it looks like scp.

	toParse := remoteURL
	if !strings.Contains(remoteURL, "://") {
		// Assume SCP-style or SSH if it has user@host:path or host:path
		// Convert "git@github.com:owner/repo" -> "ssh://git@github.com/owner/repo"
		if strings.Contains(remoteURL, ":") {
			// Replace the first : with / IF it is separating host and path
			// But we need to handle the protocol part.
			// Let's just prepend ssh:// and replace the first : with /
			// Wait, "git@host:path" -> "ssh://git@host/path"

			// Find the colon that separates host/port from path.
			// This is heuristic.
			idx := strings.Index(remoteURL, ":")
			if idx != -1 {
				toParse = "ssh://" + remoteURL[:idx] + "/" + remoteURL[idx+1:]
			}
		}
	}

	u, err := url.Parse(toParse)
	if err != nil {
		// Fallback to simple split if parsing fails?
		// Or should we error out?
		// The original code was loose.
		// Let's error out if we can't parse it safely.
		return "", "", fmt.Errorf("cannot parse generic remote: %w", err)
	}

	// Use path.Clean to resolve .. and .
	cleanedPath := path.Clean(u.Path)

	// Remove leading slash if present
	cleanedPath = strings.TrimPrefix(cleanedPath, "/")

	parts := strings.Split(cleanedPath, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("cannot parse generic remote: %s", remoteURL)
	}

	// Extract the last two parts
	owner = parts[len(parts)-2]
	repo = parts[len(parts)-1]

	// Final check: prevent ".." from slipping through if path.Clean left it (e.g. if path was ../../foo)
	if owner == ".." || repo == ".." || owner == "." || repo == "." {
		return "", "", fmt.Errorf("invalid path components in remote url: %s", remoteURL)
	}

	return owner, repo, nil
}
