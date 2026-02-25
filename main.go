package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		configPath string
		help       bool
		rate       int
		workflow   string
	)

	flag.StringVar(&configPath, "c", "", "Path to configuration file")
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.BoolVar(&help, "h", false, "Show help message and exit")
	flag.BoolVar(&help, "help", false, "Show help message and exit")
	flag.IntVar(&rate, "r", 30, "Refresh rate in seconds")
	flag.IntVar(&rate, "rate", 30, "Refresh rate in seconds")
	flag.StringVar(&workflow, "w", "Build", "GitHub Actions workflow to monitor")
	flag.StringVar(&workflow, "workflow", "Build", "GitHub Actions workflow to monitor")
	flag.Parse()

	if help {
		printUsage()
		os.Exit(0)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: GITHUB_TOKEN environment variable is not set")
		os.Exit(1)
	}

	repos := flag.Args()

	if len(repos) == 0 {
		if configPath == "" {
			configPath = DefaultConfigPath()
		}
		configRepos, err := LoadConfigFromFile(configPath)
		if err == nil && len(configRepos) > 0 {
			repos = configRepos
		}
	}

	if len(repos) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no repositories specified")
		printUsage()
		os.Exit(1)
	}

	if err := RunTUI(workflow, repos, rate, token); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: ghamon [options] [repo]...")
	fmt.Println()
	fmt.Println("GHA Monitor - Monitor GitHub Actions workflows")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -c, --config     Path to configuration file (default: $HOME/.ghamon/default)")
	fmt.Println("  -h, --help       Show help message and exit")
	fmt.Println("  -r, --rate       Refresh rate in seconds (default: 30)")
	fmt.Println("  -w, --workflow   GitHub Actions workflow to monitor (default: Build)")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  repo             GitHub repository in the format owner/repo")
}
