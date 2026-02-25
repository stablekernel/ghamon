package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromFile(t *testing.T) {
	t.Run("reads repos from config file", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "config")
		content := "owner1/repo1\nowner2/repo2\n"
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		repos, err := LoadConfigFromFile(configPath)
		require.NoError(t, err)
		assert.Equal(t, []string{"owner1/repo1", "owner2/repo2"}, repos)
	})

	t.Run("ignores comments and blank lines", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "config")
		content := "# comment\nowner1/repo1\n\n# another comment\nowner2/repo2\n"
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		repos, err := LoadConfigFromFile(configPath)
		require.NoError(t, err)
		assert.Equal(t, []string{"owner1/repo1", "owner2/repo2"}, repos)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := LoadConfigFromFile("/nonexistent/config")
		assert.Error(t, err)
	})

	t.Run("returns empty slice for empty file", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "config")
		err := os.WriteFile(configPath, []byte(""), 0644)
		require.NoError(t, err)

		repos, err := LoadConfigFromFile(configPath)
		require.NoError(t, err)
		assert.Empty(t, repos)
	})

	t.Run("trims whitespace from lines", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "config")
		content := "  owner1/repo1  \n\towner2/repo2\t\n"
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		repos, err := LoadConfigFromFile(configPath)
		require.NoError(t, err)
		assert.Equal(t, []string{"owner1/repo1", "owner2/repo2"}, repos)
	})
}
