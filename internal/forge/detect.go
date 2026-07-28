package forge

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// detectError is a stable detection sentinel. Prefer these (or fmt.Errorf %w
// wrapping them) over bare fmt.Errorf / errors.New so callers can errors.Is.
type detectError string

func (e detectError) Error() string { return string(e) }

// Detection error table. Dynamic detail is attached with fmt.Errorf %w.
const (
	ErrCouldNotLoadGitHubClient detectError = "could not load github client for url"
	ErrUnsupportedForgeOverride detectError = "unsupported forge override"
	ErrNoSupportedForge         detectError = "no supported forge detected for url"
	ErrGitHubTokenNotSet        detectError = "GITHUB_TOKEN not set"
	ErrNoRemoteURL              detectError = "could not determine remote url for 'origin' or 'upstream'"
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
			return nil, fmt.Errorf("%w: %s", ErrCouldNotLoadGitHubClient, originURL)
		default:
			return nil, fmt.Errorf("%w %q (supported: github)", ErrUnsupportedForgeOverride, overrideForge)
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

	return nil, fmt.Errorf("%w: %s", ErrNoSupportedForge, originURL)
}

// missingCredentialsError returns a clear error when the remote matches a known forge
// but GITHUB_TOKEN is unset (loaders return nil for both "not this forge" and "no token").
func missingCredentialsError(originURL string) error {
	if os.Getenv("GITHUB_TOKEN") != "" {
		return nil
	}

	if _, _, err := ParseGitHubRemote(originURL); err == nil {
		return fmt.Errorf("%w (GitHub remote detected)", ErrGitHubTokenNotSet)
	}

	// Generic Gitea/Forgejo remotes also authenticate with GITHUB_TOKEN.
	if _, _, err := ParseGenericRemote(originURL); err == nil {
		host, _ := getHostAndScheme(originURL)
		if host != "" && host != "github.com" && host != "api.github.com" {
			return fmt.Errorf("%w (forge remote detected at %s)", ErrGitHubTokenNotSet, host)
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

	return "", ErrNoRemoteURL
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
