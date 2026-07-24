package forge

import (
	"context"
	"net"
	"net/url"
	"strings"
)

// normalizeRemoteURL strips trailing "/" and optional ".git" in any order so
// forms like https://host/o/r.git/ and git@host:o/r.git/ parse the same as
// without those suffixes. A single TrimSuffix(".git") then TrimSuffix("/")
// leaves ".git" when the URL ends with ".git/".
//
// It also lowercases the host and drops default scheme ports (http:80,
// https/ssh:443). ParseGitHubRemote matches hosts with == and literal
// prefixes, so https://GitHub.com/... or https://github.com:443/... would
// otherwise fail detection and fall through as "unsupported forge".
func normalizeRemoteURL(remoteURL string) string {
	for {
		prev := remoteURL
		remoteURL = strings.TrimSuffix(remoteURL, "/")
		remoteURL = strings.TrimSuffix(remoteURL, ".git")
		if remoteURL == prev {
			break
		}
	}
	return normalizeRemoteHost(remoteURL)
}

// normalizeRemoteHost lowercases the remote host and strips default ports.
// Non-default ports (e.g. Gitea on :3000) are preserved.
func normalizeRemoteHost(remoteURL string) string {
	if remoteURL == "" {
		return remoteURL
	}

	// http(s):// and ssh:// — rewrite Host via net/url.
	if strings.Contains(remoteURL, "://") {
		u, err := url.Parse(remoteURL)
		if err != nil || u.Host == "" {
			return remoteURL
		}
		host := strings.ToLower(u.Hostname())
		port := u.Port()
		if port != "" && isDefaultRemotePort(u.Scheme, port) {
			port = ""
		}
		u.Scheme = strings.ToLower(u.Scheme)
		u.Host = joinHostPort(host, port)
		return u.String()
	}

	// SCP-like: user@host:path (no scheme). Lowercase host only.
	at := strings.LastIndex(remoteURL, "@")
	if at < 0 {
		return remoteURL
	}
	rest := remoteURL[at+1:]
	colon := strings.Index(rest, ":")
	if colon <= 0 {
		return remoteURL
	}
	host := strings.ToLower(rest[:colon])
	return remoteURL[:at+1] + host + rest[colon:]
}

// isDefaultRemotePort reports whether port is the scheme's default and can be
// omitted without changing the remote identity.
func isDefaultRemotePort(scheme, port string) bool {
	switch strings.ToLower(scheme) {
	case "http":
		return port == "80"
	case "https", "ssh":
		return port == "443"
	default:
		return false
	}
}

// joinHostPort builds a URL host (host or host:port), bracketing IPv6 hosts.
func joinHostPort(host, port string) string {
	if port == "" {
		if strings.Contains(host, ":") {
			return "[" + host + "]"
		}
		return host
	}
	return net.JoinHostPort(host, port)
}

// State represents the status of a CI/CD pipeline step reported to the forge.
// Different forges might map these states slightly differently (e.g. GitHub treats 'running' as 'pending').
type State string

const (
	// StateRunning indicates the task is currently executing.
	// Note: GitHub API maps this to 'pending' with a description unless using check runs.
	StateRunning State = "running"
	// StatePending indicates the task is queued or waiting.
	StatePending State = "pending"
	// StateSuccess indicates the task completed successfully (exit code 0).
	StateSuccess State = "success"
	// StateFailure indicates the task failed (non-zero exit code).
	StateFailure State = "failure"
	// StateError indicates a configuration or runtime error prevented the task from running properly.
	StateError   State = "error"
)

// StatusOpts encapsulates the parameters required to set a commit status.
type StatusOpts struct {
	// Commit is the SHA-1 hash of the commit to update.
	Commit      string
	// Context is the label that differentiates this status check (e.g., "ci/lint").
	Context     string
	// State is the current status of the task.
	State       State
	// Description is a short, human-readable summary of the status.
	Description string
	// TargetURL is an optional link to the build details (e.g., CI logs).
	TargetURL   string
}

// ForgeClient defines the interface for interacting with a Git forge (GitHub, GitLab, Gitea, etc.).
// Implementations handle the specifics of authentication and API calls.
type ForgeClient interface {
	// SetStatus updates the commit status for the given options.
	// It should handle API-specific nuances, such as state mapping.
	SetStatus(ctx context.Context, opts StatusOpts) error
}

// ForgeLoader is a strategy function that attempts to instantiate a ForgeClient from a remote URL.
// It returns nil if the URL is not supported by this strategy, allowing the next strategy to be tried.
type ForgeLoader func(url string) ForgeClient
