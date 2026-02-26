# GHA Monitor (ghamon)

GHA Monitor (ghamon) is a TUI program that monitors a given GitHub Actions workflow for one or more GitHub repositories.


## Usage

```bash
ghamon [options] [repo]...
```

Options:

- -c (--config) -- Path to configuration file (default: $HOME/.ghamon/default)
- -h (--help) -- Show help message and exit
- -r (--rate) -- Refresh rate in seconds (default: 30 seconds)
- -w (--workflow) -- GitHub Actions workflow to monitor (default: all workflows)

Arguments:

- repo (optional) -- GitHub repository in the format `owner/repo`; multiple can be specified


## Configuration

The configuration file is a simple text file with one repository per line. Lines starting with `#` are treated as comments and ignored.


## Design

### User Interface

CLI for invocation, TUI for data presentation.

#### TUI Layout

The TUI layout consists of a header, a main content area, and a footer. The header displays the title, workflow name, refresh rate, and a progress bar. The content area is divided into columns for repository and status. The footer provides instructions for quitting the application and refreshing the data manually.

The full terminal area is used. The content area is scrollable if the number of rows exceeds the available space. The header and footer remain fixed at the top and bottom of the screen, respectively, while the main content area scrolls independently.

An alternate display buffer is used, allowing the application to take full control of the terminal display without affecting the normal terminal output. This ensures that when the application exits, the terminal is restored to its original state without any residual output from the TUI.

The progress bar shown in the header is refreshed during data retrieval to provide visual feedback to the user that the application is actively fetching data from the GitHub API. The progress bar is cleared when all data has been retrieved and the display is updated with the latest workflow status information.

Workflows are displayed in the order they are defined in the configuration file. If no configuration file is provided, all workflows for the specified repositories are monitored and displayed in the order returned by the GitHub API. Workflow names that start with ".github/workflows/" are displayed without the prefix for better readability. Workflows that start with "Graph Update" or "go_modules" are not displayed.

#### Key Bindings

- `q` -- Quit the application
- `r` -- Refresh the data manually

### Data Retrieval

Data is retrieved from the GitHub API at the specified refresh rate. Credentials for accessing the GitHub API must be specified in the environment variable `GITHUB_TOKEN`. The user is assumed to be x-oauth-basic.

### Technical Constraints

* Application implemented in Go.
* Build and test automated with Mage. Targets show command output.
* Testify packages used for assertions and mock in unit tests.
* Usage information printed to stdout.


## Build

To build the application, run the following command:

```bash
mage build
```

## Test

To run the unit tests, execute:

```bash
mage test
```
