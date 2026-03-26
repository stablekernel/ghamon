package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultConfigPath returns the default configuration file path.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ghamon", "default")
}

// Load reads a configuration file and returns the list of repositories.
// Lines beginning with '#' are treated as comments and ignored.
// Empty lines are also ignored.
func Load(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening config file %q: %w", path, err)
	}
	defer f.Close()

	var repos []string
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := validateRepo(line); err != nil {
			return nil, fmt.Errorf("config file %q line %d: %w", path, lineNum, err)
		}
		repos = append(repos, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}
	return repos, nil
}

// validateRepo checks that the repository string is in "owner/repo" format.
func validateRepo(s string) error {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid repository %q: must be in owner/repo format", s)
	}
	return nil
}
