package forge

import (
	"fmt"
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
	u, err := ParseRemoteURL(remoteURL)
	if err != nil {
		return nil
	}
	host := u.Host
	scheme := u.Scheme

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

	// If scheme is SSH, assume HTTPS for API
	if scheme == "ssh" {
		scheme = "https"
	}

	client.BaseURL = fmt.Sprintf("%s://%s/api/v1", scheme, host)

	return client
}

// ParseGenericRemote extracts the owner and repository from a generic remote URL.
// It uses ParseRemoteURL to normalize and parse the URL, ensuring security and consistency.
func ParseGenericRemote(remoteURL string) (owner, repo string, err error) {
	u, err := ParseRemoteURL(remoteURL)
	if err != nil {
		return "", "", err
	}

	// Remove leading slash from path
	repoPath := strings.TrimPrefix(u.Path, "/")

	parts := strings.Split(repoPath, "/")
	// Filter empty parts if any (e.g. double slashes)
	var cleanParts []string
	for _, p := range parts {
		if p != "" {
			cleanParts = append(cleanParts, p)
		}
	}

	if len(cleanParts) < 2 {
		return "", "", fmt.Errorf("cannot parse generic remote: %s", remoteURL)
	}

	// Taking the last two parts handles cases like ssh://host/group/subgroup/owner/repo
	return cleanParts[len(cleanParts)-2], cleanParts[len(cleanParts)-1], nil
}
