package forge

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DetectClient attempts to detect and initialize a ForgeClient based on the repository remote URL.
// It iterates through available strategies (GitHub, Generic).
func DetectClient(overrideForge string) (ForgeClient, error) {
	// If forge is explicitly set to github, we try GitHub strategy only?
	// The override logic in original code was checking "github" -> ParseGitHubRemote.

	// First get the remote URL.
	// For "GitHub" override, we still need the URL to get owner/repo unless we want to force something.
	// But `DetectRepoInfo` previously did parsing.

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

func getOriginURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}

	// Fallback to checking first remote if origin fails?
    cmd = exec.Command("git", "remote", "-v")
    out, err = cmd.Output()
    if err != nil {
        return "", err
    }

    lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		fields := strings.Fields(lines[0])
		if len(fields) >= 2 {
			return fields[1], nil
		}
	}
	return "", fmt.Errorf("could not determine remote url")
}

// Deprecated: Logic moved to DetectClient and strategies
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

// Deprecated: Use DetectClient strategies instead
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
