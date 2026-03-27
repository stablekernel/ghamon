# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

GHA Monitor (ghamon) is a TUI application that monitors GitHub Actions workflow status for one or more GitHub repositories. It uses the Bubble Tea framework for the terminal UI and lipgloss for styling.

## Build & Test

Uses [Mage](https://magefile.org/) as the build tool (see `magefile.go`):

```bash
mage build    # builds ./bin/ghamon binary
mage test     # runs go test -v ./...
mage clean    # removes ./bin directory
```

Run a single test:
```bash
go test -v -run <name> ./...
```

## Architecture

The project follows a standard Go layout with two packages:

- **cmd/main.go** — CLI entry point (`package main`): parses flags (`-r`/`-w`/`-h`), reads `GITHUB_TOKEN` env var, resolves repos, launches TUI. Imports `ghamon/internal/ghamon`.
- **internal/ghamon/** — Core library (`package ghamon`) with three source files:
  - **repos.go** — Repository resolution: handles `owner/repo` args, `@file` references (reads from `~/.ghamon/`), and auto-detection of current git repo's GitHub remote.
  - **github.go** — GitHub API client: `FetchWorkflowRuns` (all workflows, deduplicated by `workflow_id`) and `FetchWorkflowRun` (single named workflow). Uses `httptest` servers in tests for HTTP mocking.
  - **tui.go** — Bubble Tea model: manages fetch sequencing (one repo at a time via `fetchNextMsg`/`fetchedRepoMsg` message chain), scrollable content area, progress bar, and status animation for in-progress workflows. Filters out "Graph Update" and "go_modules" workflows.

## Coding Conventions

* Always use `:=` instead of var in functions.
* Keep functions small and focused on a single responsibility.

## Testing Conventions

- Tests use `testify/assert` and `testify/require` for assertions.
- GitHub API tests use `httptest.NewServer` with a `GitHubClient` pointed at the test server's URL.
- Table-driven tests are used where appropriate (see `TestParseGitHubRepo`).

## Runtime Requirements

- `GITHUB_TOKEN` environment variable must be set.
- Repository lists can be stored in `~/.ghamon/<name>` and referenced with `@<name>`.
