package forge

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseRemoteURL parses a git remote URL into a standard *url.URL.
// It handles standard URLs (http, https, ssh) and SCP-style URLs (git@host:path).
// SCP-style URLs are converted to ssh://git@host/path before parsing.
// The .git suffix is removed from the path, if present.
func ParseRemoteURL(rawURL string) (*url.URL, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("remote url cannot be empty")
	}

	// Remove .git suffix if present
	rawURL = strings.TrimSuffix(rawURL, ".git")

	// Check if it's an SCP-style URL (git@host:path)
	// Must have @, :, and no scheme (://)
	if strings.Contains(rawURL, "@") && strings.Contains(rawURL, ":") && !strings.Contains(rawURL, "://") {
		// Convert git@host:path -> ssh://git@host/path
		// We replace the FIRST colon that acts as the separator.
		colonIdx := strings.Index(rawURL, ":")
		if colonIdx != -1 {
			rawURL = "ssh://" + rawURL[:colonIdx] + "/" + rawURL[colonIdx+1:]
		}
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote url: %w", err)
	}

	return u, nil
}
