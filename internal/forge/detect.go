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
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err == nil {
		url := strings.TrimSpace(string(out))
		owner, repo, err := ParseGitHubRemote(url)
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

	return "", "", fmt.Errorf("could not detect repo info")
}
