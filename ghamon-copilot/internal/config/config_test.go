package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ghamon/internal/config"
)

func TestLoad_MissingFile(t *testing.T) {
	repos, err := config.Load("/nonexistent/path/to/config")
	require.NoError(t, err)
	assert.Nil(t, repos)
}

func TestLoad_EmptyFile(t *testing.T) {
	f := writeTempConfig(t, "")
	repos, err := config.Load(f)
	require.NoError(t, err)
	assert.Empty(t, repos)
}

func TestLoad_CommentsAndBlanks(t *testing.T) {
	content := "# This is a comment\n   # Indented comment\n\n"
	f := writeTempConfig(t, content)
	repos, err := config.Load(f)
	require.NoError(t, err)
	assert.Empty(t, repos)
}

func TestLoad_ValidRepos(t *testing.T) {
	content := "# My repos\nowner/repo1\nowner/repo2\nanother-org/another-repo\n"
	f := writeTempConfig(t, content)
	repos, err := config.Load(f)
	require.NoError(t, err)
	assert.Equal(t, []string{"owner/repo1", "owner/repo2", "another-org/another-repo"}, repos)
}

func TestLoad_InvalidRepo(t *testing.T) {
	f := writeTempConfig(t, "not-a-valid-repo\n")
	_, err := config.Load(f)
	assert.Error(t, err)
}

func TestLoad_InvalidRepo_EmptyOwner(t *testing.T) {
	f := writeTempConfig(t, "/repo\n")
	_, err := config.Load(f)
	assert.Error(t, err)
}

func TestLoad_InvalidRepo_EmptyName(t *testing.T) {
	f := writeTempConfig(t, "owner/\n")
	_, err := config.Load(f)
	assert.Error(t, err)
}

func TestDefaultConfigPath(t *testing.T) {
	p := config.DefaultConfigPath()
	assert.NotEmpty(t, p)
	assert.Contains(t, p, ".ghamon")
	assert.True(t, filepath.IsAbs(p))
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}
