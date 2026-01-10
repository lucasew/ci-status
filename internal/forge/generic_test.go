package forge_test

import (
	"testing"

	"ci-status/internal/forge"
)

func TestParseGenericRemote(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		owner   string
		repo    string
		wantErr bool
	}{
		{
			name:    "HTTPS URL",
			url:     "https://gitea.com/owner/repo.git",
			owner:   "owner",
			repo:    "repo",
			wantErr: false,
		},
		{
			name:    "SSH URL",
			url:     "ssh://git@gitea.com/owner/repo",
			owner:   "owner",
			repo:    "repo",
			wantErr: false,
		},
		{
			name:    "SCP-like URL",
			url:     "git@gitea.com:owner/repo.git",
			owner:   "owner",
			repo:    "repo",
			wantErr: false,
		},
		{
			name:    "URL with no path",
			url:     "https://gitea.com/",
			owner:   "",
			repo:    "",
			wantErr: true,
		},
		{
			name:    "URL with only one path component",
			url:     "https://gitea.com/owner",
			owner:   "",
			repo:    "",
			wantErr: true,
		},
		{
			name:    "Malformed URL that could cause path injection",
			url:     "https://gitea.com",
			owner:   "",
			repo:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := forge.ParseGenericRemote(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGenericRemote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if owner != tt.owner {
				t.Errorf("ParseGenericRemote() owner = %v, want %v", owner, tt.owner)
			}
			if repo != tt.repo {
				t.Errorf("ParseGenericRemote() repo = %v, want %v", repo, tt.repo)
			}
		})
	}
}
