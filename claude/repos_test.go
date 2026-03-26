package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGitHubRepo(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		want      string
		wantErr   bool
	}{
		{
			name:      "SSH URL",
			remoteURL: "git@github.com:owner/repo.git",
			want:      "owner/repo",
		},
		{
			name:      "HTTPS URL with .git",
			remoteURL: "https://github.com/owner/repo.git",
			want:      "owner/repo",
		},
		{
			name:      "HTTPS URL without .git",
			remoteURL: "https://github.com/owner/repo",
			want:      "owner/repo",
		},
		{
			name:      "non-GitHub remote",
			remoteURL: "https://gitlab.com/owner/repo.git",
			wantErr:   true,
		},
		{
			name:      "malformed URL",
			remoteURL: "git@github.com:",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGitHubRepo(tt.remoteURL)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestLoadReposFromFile(t *testing.T) {
	t.Run("reads repos from file", func(t *testing.T) {
		path := writeTempFile(t, "owner1/repo1\nowner2/repo2\n")
		repos, err := LoadReposFromFile(path)
		require.NoError(t, err)
		assert.Equal(t, []string{"owner1/repo1", "owner2/repo2"}, repos)
	})

	t.Run("ignores comments and blank lines", func(t *testing.T) {
		path := writeTempFile(t, "# comment\nowner1/repo1\n\n# another\nowner2/repo2\n")
		repos, err := LoadReposFromFile(path)
		require.NoError(t, err)
		assert.Equal(t, []string{"owner1/repo1", "owner2/repo2"}, repos)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		path := writeTempFile(t, "  owner1/repo1  \n\towner2/repo2\t\n")
		repos, err := LoadReposFromFile(path)
		require.NoError(t, err)
		assert.Equal(t, []string{"owner1/repo1", "owner2/repo2"}, repos)
	})

	t.Run("returns empty slice for empty file", func(t *testing.T) {
		path := writeTempFile(t, "")
		repos, err := LoadReposFromFile(path)
		require.NoError(t, err)
		assert.Empty(t, repos)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := LoadReposFromFile("/nonexistent/file")
		assert.Error(t, err)
	})
}

func TestResolveRepos(t *testing.T) {
	t.Run("returns direct repo args", func(t *testing.T) {
		repos, err := ResolveRepos([]string{"owner1/repo1", "owner2/repo2"})
		require.NoError(t, err)
		assert.Equal(t, []string{"owner1/repo1", "owner2/repo2"}, repos)
	})

	t.Run("expands @file argument", func(t *testing.T) {
		path := writeTempFile(t, "owner1/repo1\nowner2/repo2\n")
		repos, err := ResolveRepos([]string{"@" + path})
		require.NoError(t, err)
		assert.Equal(t, []string{"owner1/repo1", "owner2/repo2"}, repos)
	})

	t.Run("mixes direct repos and @file", func(t *testing.T) {
		path := writeTempFile(t, "owner2/repo2\n")
		repos, err := ResolveRepos([]string{"owner1/repo1", "@" + path})
		require.NoError(t, err)
		assert.Equal(t, []string{"owner1/repo1", "owner2/repo2"}, repos)
	})

	t.Run("returns error for missing @file", func(t *testing.T) {
		_, err := ResolveRepos([]string{"@/nonexistent/file"})
		assert.Error(t, err)
	})
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "repos")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}
