# GHA Monitor (ghamon)

GHA Monitor (ghamon) is a TUI program that monitors GitHub Actions workflows for one or more GitHub repositories.


## Usage

```bash
ghamon [options] [repo]...
```

Run `ghamon --help` for more information on available options and usage instructions.


## Configuration

The configuration file is a simple text file with one repository per line. Lines starting with `#` are treated as comments and ignored.


## Build

To build the application:

```bash
mage build
```


## Test

To run the unit tests:

```bash
mage test
```
