package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	flag "github.com/spf13/pflag"

	"ghamon/internal/config"
	ghclient "ghamon/internal/github"
	"ghamon/internal/tui"
)

const defaultRate = 30

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("ghamon", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var (
		configPath string
		rate       int
		workflow   string
		showHelp   bool
	)

	fs.StringVarP(&configPath, "config", "c", config.DefaultConfigPath(), "Path to configuration file")
	fs.IntVarP(&rate, "rate", "r", defaultRate, "Refresh rate in seconds")
	fs.StringVarP(&workflow, "workflow", "w", "", "GitHub Actions workflow to monitor (default: all workflows)")
	fs.BoolVarP(&showHelp, "help", "h", false, "Show help message and exit")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if showHelp {
		printUsage(fs)
		return nil
	}

	cfgRepos, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
	}

	repos := dedupe(append(cfgRepos, fs.Args()...))

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	if rate < 1 {
		rate = defaultRate
	}

	client := ghclient.New(token)
	model := tui.New(repos, workflow, rate, client)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}
	return nil
}

func printUsage(fs *flag.FlagSet) {
	fmt.Println("Usage: ghamon [options] [repo]...")
	fmt.Println()
	fmt.Println("GHA Monitor monitors GitHub Actions workflows for one or more repositories.")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  repo    GitHub repository in the format owner/repo (multiple allowed)")
	fmt.Println()
	fmt.Println("Options:")
	fs.PrintDefaults()
}

// dedupe removes duplicate repositories while preserving insertion order.
func dedupe(repos []string) []string {
	seen := make(map[string]bool, len(repos))
	out := make([]string, 0, len(repos))
	for _, r := range repos {
		key := strings.ToLower(r)
		if !seen[key] {
			seen[key] = true
			out = append(out, r)
		}
	}
	return out
}
