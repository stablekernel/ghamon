---
output: ghamon.go
package: ghamon
---
# Run()

- Call ParseArgs() to get repositories, workflow, and refreshRate.
- If repositories is empty: Call LoadConfig() to get repositories.
- Call tui.Run(repositories, workflow, refreshRate) to start the TUI.

# ParseArgs() -> repositories, workflow, refreshRate

- Define command line flags.
- Call flag.Parse().
- Get workflow from -w flag argument.
- Get refreshRate from -r flag argument.
- Get repositories from arguments.
- Return repositories, workflow, and refreshRate.

# LoadConfig() -> repositories

- Open configuration file.
- For each line in file:
  - If line starts with `#` (comment marker) or is blank: Skip it.
  - Add trimmed line to repositories.
- Return repositories.
