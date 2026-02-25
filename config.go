package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// DefaultConfigPath returns the default configuration file path: $HOME/.ghamon/default.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".ghamon", "default")
	}
	return filepath.Join(home, ".ghamon", "default")
}

// LoadConfigFromFile loads repository names from the given file path.
// Each line should contain one repository in owner/repo format.
// Lines starting with # are treated as comments and ignored.
func LoadConfigFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var repos []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		repos = append(repos, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return repos, nil
}
