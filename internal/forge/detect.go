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
//  1. Retrieves the 'origin' or 'upstream' remote URL.
//  2. If 'overrideForge' is set (e.g. "github"), only that strategy is used; unknown
//     overrides error without falling through to auto-detect.
//  3. Otherwise, it iterates through all registered strategies in precedence order.
//  4. If a known forge remote is present but credentials are missing, returns a credentials error
//     instead of the generic "no supported forge" message.
//
// Returns:
// - ForgeClient: An initialized client ready for API calls.
// - error: If no supported forge is detected, credentials are missing, or remote URL retrieval fails.
func DetectClient(overrideForge string) (ForgeClient, error) {
	originURL, err := getOriginURL()
	if err != nil {
		return nil, err
	}
	return detectClientFromURL(originURL, overrideForge)
}

// detectClientFromURL selects a ForgeClient for a remote URL.
// Extracted so unit tests can cover credential vs. unsupported-host errors without a git repo.
//
// When overrideForge is set, only that strategy is used (no auto-detect fallthrough).
// Unknown overrides fail immediately so typos do not silently report to another forge.
func detectClientFromURL(originURL, overrideForge string) (ForgeClient, error) {
	if overrideForge != "" {
		switch overrideForge {
		case "github":
			if client := LoadGitHub(originURL); client != nil {
				return client, nil
			}
			if err := missingCredentialsError(originURL); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("could not load github client for url: %s", originURL)
		default:
			return nil, fmt.Errorf("unsupported forge override %q (supported: github)", overrideForge)
		}
	}

	// Auto-detect: try strategies in order of precedence.
	strategies := []ForgeLoader{
		LoadGitHub,
		LoadGeneric,
	}

	for _, strategy := range strategies {
		if client := strategy(originURL); client != nil {
			return client, nil
		}
	}

	if err := missingCredentialsError(originURL); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("no supported forge detected for url: %s", originURL)
}

// missingCredentialsError returns a clear error when the remote matches a known forge
// but GITHUB_TOKEN is unset (loaders return nil for both "not this forge" and "no token").
func missingCredentialsError(originURL string) error {
	if os.Getenv("GITHUB_TOKEN") != "" {
		return nil
	}

	if _, _, err := ParseGitHubRemote(originURL); err == nil {
		return fmt.Errorf("GITHUB_TOKEN not set (GitHub remote detected)")
	}

	// Generic Gitea/Forgejo remotes also authenticate with GITHUB_TOKEN.
	if _, _, err := ParseGenericRemote(originURL); err == nil {
		host, _ := getHostAndScheme(originURL)
		if host != "" && host != "github.com" && host != "api.github.com" {
			return fmt.Errorf("GITHUB_TOKEN not set (forge remote detected at %s)", host)
		}
	}

	return nil
}

// getOriginURL retrieves the remote URL for the repository.
// It attempts to read from the 'origin' remote first, falling back to 'upstream' if 'origin' is not defined.
// This supports forked repositories where the upstream might be the primary source of truth.
func getOriginURL() (string, error) {
	for _, remote := range []string{"origin", "upstream"} {
		cmd := exec.Command("git", "remote", "get-url", remote)
		out, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}

	return "", fmt.Errorf("could not determine remote url for 'origin' or 'upstream'")
}

// DetectCommit resolves the commit SHA to be reported.
// It prioritizes the override value, then CI environment variables (GITHUB_SHA, CI_COMMIT_SHA, BITBUCKET_COMMIT),
// and finally falls back to the current git HEAD.
func DetectCommit(override string) (string, error) {
	if override != "" {
		return override, nil
	}

	// CI Env vars
	for _, env := range []string{"GITHUB_SHA", "CI_COMMIT_SHA", "BITBUCKET_COMMIT"} {
		if sha := os.Getenv(env); sha != "" {
			return sha, nil
		}
	}

	// Git fallback
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
