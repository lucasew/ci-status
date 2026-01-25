package forge

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DetectClient attempts to detect and initialize a ForgeClient based on the repository remote URL.
//
// It uses a strategy pattern, iterating through available loaders (GitHub, Generic)
// until one successfully claims the URL.
//
// If 'overrideForge' is provided (e.g. "github"), it prioritizes that strategy.
func DetectClient(overrideForge string) (ForgeClient, error) {
	originURL, err := getOriginURL()
	if err != nil {
		return nil, err
	}

	if overrideForge == "github" {
		client := LoadGitHub(originURL)
		if client != nil {
			return client, nil
		}
		// If explicit override but failed to load (e.g. format mismatch), we might error out.
		// But let's stick to the flow.
	}

	// Try strategies in order
	strategies := []ForgeLoader{
		LoadGitHub,
		LoadGeneric,
	}

	for _, strategy := range strategies {
		if client := strategy(originURL); client != nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("no supported forge detected for url: %s", originURL)
}

// getOriginURL retrieves the remote URL from git.
//
// It first attempts to get the URL for the 'origin' remote.
// If that fails (e.g. detached HEAD or no origin), it falls back to the 'upstream' remote.
func getOriginURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}

	// Fallback to checking 'upstream' remote if 'origin' fails
	cmd = exec.Command("git", "remote", "get-url", "upstream")
	out, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}

	return "", fmt.Errorf("could not determine remote url for 'origin' or 'upstream'")
}

// DetectForge attempts to identify the forge type from git remotes.
//
// Deprecated: This logic has been moved to DetectClient and specific ForgeLoaders.
// New code should use DetectClient instead.
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

// DetectCommit determines the commit SHA to report status for.
//
// It resolves the commit in the following order of precedence:
// 1. Explicit override (CLI flag)
// 2. CI environment variables (GITHUB_SHA, CI_COMMIT_SHA, BITBUCKET_COMMIT)
// 3. Local git HEAD (via 'git rev-parse HEAD')
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

// DetectURL returns the target URL for the status description.
// Currently only supports explicit override via CLI flag.
func DetectURL(override string) string {
    if override != "" {
        return override
    }
    // No explicit detection logic in spec for URL other than flag
    return ""
}

// DetectRepoInfo attempts to parse owner and repo from the remote URL.
//
// Deprecated: Use DetectClient strategies instead. This function relies on
// specific forge implementation details that should be encapsulated.
func DetectRepoInfo() (string, string, error) {
	// This function is kept for backward compatibility if needed,
    // but the implementation logic is now in strategies.
    // We can implement it using ParseGenericRemote as a fallback wrapper.
	originURL, err := getOriginURL()
	if err != nil {
		return "", "", err
	}

	owner, repo, err := ParseGitHubRemote(originURL)
	if err == nil {
		return owner, repo, nil
	}

	return ParseGenericRemote(originURL)
}
