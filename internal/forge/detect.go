package forge

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DetectClient attempts to identify the appropriate ForgeClient by analyzing the repository's remote URL.
// It implements a strategy pattern, iterating through available loaders (GitHub, Generic) to find a match.
//
// Behavior:
// 1. Retrieves the 'origin' or 'upstream' remote URL.
// 2. If 'overrideForge' is set (e.g. "github"), it tries that specific strategy first.
// 3. Otherwise, it iterates through all registered strategies in precedence order.
//
// Returns:
// - ForgeClient: An initialized client ready for API calls.
// - error: If no supported forge is detected or if remote URL retrieval fails.
func DetectClient(overrideForge string) (ForgeClient, error) {
	originURL, err := getOriginURL()
	if err != nil {
		return nil, err
	}

	// If the user explicitly requested GitHub, try that strategy first.
	if overrideForge == "github" {
		client := LoadGitHub(originURL)
		if client != nil {
			return client, nil
		}
	}

	// Iterate through supported strategies in order of precedence.
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

// getOriginURL retrieves the remote URL for the repository.
// It attempts to read from the 'origin' remote first, falling back to 'upstream' if 'origin' is not defined.
// This supports forked repositories where the upstream might be the primary source of truth.
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

// DetectForge identifies the forge type from the git remote URL or an override.
// @deprecated: Logic moved to DetectClient and specific ForgeLoader strategies.
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

// DetectCommit resolves the commit SHA to be reported.
// It prioritizes the override value, then CI environment variables (GITHUB_SHA, CI_COMMIT_SHA, BITBUCKET_COMMIT),
// and finally falls back to the current git HEAD.
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
// Currently, it acts as a passthrough for the override value, as there is no automatic detection logic implemented.
func DetectURL(override string) string {
	if override != "" {
		return override
	}
	return ""
}

// DetectRepoInfo attempts to extract the owner and repository name from the remote URL.
// @deprecated: Use DetectClient strategies instead. This function is retained for backward compatibility.
func DetectRepoInfo() (string, string, error) {
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
