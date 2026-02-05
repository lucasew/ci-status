package forge_test

import (
	"testing"

	"ci-status/internal/forge"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		wantHost string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "HTTPS URL",
			rawURL:   "https://github.com/owner/repo.git",
			wantHost: "github.com",
			wantPath: "/owner/repo",
			wantErr:  false,
		},
		{
			name:     "HTTPS URL without suffix",
			rawURL:   "https://github.com/owner/repo",
			wantHost: "github.com",
			wantPath: "/owner/repo",
			wantErr:  false,
		},
		{
			name:     "SSH URL",
			rawURL:   "ssh://git@github.com/owner/repo.git",
			wantHost: "github.com",
			wantPath: "/owner/repo",
			wantErr:  false,
		},
		{
			name:     "SCP-style URL",
			rawURL:   "git@github.com:owner/repo.git",
			wantHost: "github.com",
			wantPath: "/owner/repo",
			wantErr:  false,
		},
		{
			name:     "SCP-style URL without suffix",
			rawURL:   "git@github.com:owner/repo",
			wantHost: "github.com",
			wantPath: "/owner/repo",
			wantErr:  false,
		},
		{
			name:     "Generic HTTPS",
			rawURL:   "https://git.example.com/group/subgroup/project",
			wantHost: "git.example.com",
			wantPath: "/group/subgroup/project",
			wantErr:  false,
		},
		{
			name:     "Empty URL",
			rawURL:   "",
			wantHost: "",
			wantPath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := forge.ParseRemoteURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRemoteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Host != tt.wantHost {
					t.Errorf("ParseRemoteURL() host = %v, want %v", got.Host, tt.wantHost)
				}
				if got.Path != tt.wantPath {
					t.Errorf("ParseRemoteURL() path = %v, want %v", got.Path, tt.wantPath)
				}
			}
		})
	}
}
