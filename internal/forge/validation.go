package forge

import (
	"fmt"
	"regexp"
)

// validRepoSegment enforces strict naming for repository owners and names.
// It allows alphanumeric characters, underscores, hyphens, and periods.
// It explicitly rejects control characters, slashes, backslashes, query parameters,
// and path traversal sequences like "." and "..".
var validRepoSegment = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// validateRepoName checks if a repository owner or name segment is valid.
func validateRepoName(segment string) error {
	if segment == "" {
		return fmt.Errorf("repository segment cannot be empty")
	}
	if segment == "." || segment == ".." {
		return fmt.Errorf("repository segment cannot be '.' or '..'")
	}
	if !validRepoSegment.MatchString(segment) {
		return fmt.Errorf("repository segment contains invalid characters: %q", segment)
	}
	return nil
}
