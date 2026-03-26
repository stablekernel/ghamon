package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ResolveRepos resolves repository arguments into a list of owner/repo strings.
// Each arg may be owner/repo or @file (a file containing repos, one per line).
// If args is empty, the current git repository's GitHub remote is used.
func ResolveRepos(args []string) ([]string, error) {
	if len(args) == 0 {
		repo, err := CurrentGitHubRepo()
		if err != nil {
			return nil, fmt.Errorf("no repositories specified and current directory is not a git repository")
		}
		return []string{repo}, nil
	}

	var repos []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "@") {
			path := expandHome(arg[1:])
			fileRepos, err := LoadReposFromFile(path)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", path, err)
			}
			repos = append(repos, fileRepos...)
		} else {
			repos = append(repos, arg)
		}
	}
	return repos, nil
}

// CurrentGitHubRepo returns the owner/repo of the current directory's GitHub remote.
func CurrentGitHubRepo() (string, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	return parseGitHubRepo(strings.TrimSpace(string(out)))
}

// parseGitHubRepo extracts owner/repo from a GitHub remote URL.
func parseGitHubRepo(remoteURL string) (string, error) {
	var path string
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		path = strings.TrimPrefix(remoteURL, "git@github.com:")
	} else if idx := strings.Index(remoteURL, "github.com/"); idx >= 0 {
		path = remoteURL[idx+len("github.com/"):]
	} else {
		return "", fmt.Errorf("not a GitHub remote: %s", remoteURL)
	}
	path = strings.TrimSuffix(path, ".git")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("unexpected GitHub remote format: %s", remoteURL)
	}
	return parts[0] + "/" + parts[1], nil
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home + path[1:]
	}
	return path
}

// LoadReposFromFile reads repository names from a file, one per line.
// Lines starting with # are treated as comments and ignored.
func LoadReposFromFile(path string) ([]string, error) {
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
	return repos, scanner.Err()
}
