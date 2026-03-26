package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		help     bool
		rate     int
		workflow string
	)

	flag.BoolVar(&help, "h", false, "Show help message and exit")
	flag.BoolVar(&help, "help", false, "Show help message and exit")
	flag.IntVar(&rate, "r", 30, "Refresh rate in seconds")
	flag.IntVar(&rate, "rate", 30, "Refresh rate in seconds")
	flag.StringVar(&workflow, "w", "", "GitHub Actions workflow to monitor (default: all)")
	flag.StringVar(&workflow, "workflow", "", "GitHub Actions workflow to monitor (default: all)")
	flag.Usage = printUsage
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

	repos, err := ResolveRepos(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		printUsage()
		os.Exit(1)
	}

	if err := RunTUI(workflow, repos, rate, token); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: ghamon [options] [repository]...")
	fmt.Println()
	fmt.Println("GHA Monitor - Monitor GitHub Actions workflows")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help       Show help message and exit")
	fmt.Println("  -r, --rate       Refresh rate in seconds (default: 30)")
	fmt.Println("  -w, --workflow   GitHub Actions workflow to monitor (default: all)")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  repository       owner/repo, or @file (file in ~/.ghamon) containing")
	fmt.Println("                   repos one per line (default: current git repository)")
}
