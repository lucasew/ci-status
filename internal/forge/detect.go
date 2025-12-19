package forge

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func DetectForge(override string) (string, error) {
	if override != "" {
		return override, nil
	}

	cmd := exec.Command("git", "remote", "-v")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	output := string(out)

	if strings.Contains(output, "github.com") {
		return "github", nil
	}
	if strings.Contains(output, "gitlab.com") {
		return "gitlab", nil
	}
	if strings.Contains(output, "bitbucket.org") {
		return "bitbucket", nil
	}

	return "", fmt.Errorf("could not detect forge")
}

func DetectCommit(override string) (string, error) {
	if override != "" {
		return override, nil
	}

	// CI Env vars
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		return sha, nil
	}
	if sha := os.Getenv("CI_COMMIT_SHA"); sha != "" {
		return sha, nil
	}
	if sha := os.Getenv("BITBUCKET_COMMIT"); sha != "" {
		return sha, nil
	}

	// Git fallback
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func DetectURL(override string) string {
	if override != "" {
		return override
	}
	// No explicit detection logic in spec for URL other than flag
	return ""
}

func DetectRepoInfo() (string, string, error) {
	var originURL string
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err == nil {
		originURL = strings.TrimSpace(string(out))
		owner, repo, err := ParseGitHubRemote(originURL)
		if err == nil {
			return owner, repo, nil
		}
	}

	// Fallback to checking all remotes if 'origin' fails or parses incorrectly
	cmd = exec.Command("git", "remote", "-v")
	out, err = cmd.Output()
	if err != nil {
		return "", "", err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "github.com") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				owner, repo, err := ParseGitHubRemote(fields[1])
				if err == nil {
					return owner, repo, nil
				}
			}
		}
	}

	// Final Fallback: if we found an origin URL but it wasn't a standard GitHub one, try generic parsing
	if originURL != "" {
		owner, repo, err := ParseGenericRemote(originURL)
		if err == nil {
			return owner, repo, nil
		}
	}

	return "", "", fmt.Errorf("could not detect repo info")
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
