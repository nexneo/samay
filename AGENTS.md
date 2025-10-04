# Gemini Code Understanding

## Project Overview

This project, "Samay," is a terminal-first time tracking tool written in Go. It's designed for developers who are comfortable working in the terminal. The name "Samay" is the Hindi word for "Time."

The tool allows users to start and stop timers for different projects, manually log time entries, and review project logs inside the TUI. Data is stored locally in files, and the application can detect and use a Dropbox folder for data synchronization.

The project uses Google's Protocol Buffers for data serialization and the `bubbletea` library for its terminal user interface (TUI).

## Building and Running

### Building the application

To build the application, run the following command:

```sh
go build
```

This will create an executable file named `samay` in the project's root directory.

### Running the application

Running the `samay` binary with no arguments launches the Bubble Tea terminal user interface. The legacy command-line subcommands (e.g., `start`, `stop`, `report`, `web`) are currently unreachable because the application exits after starting the TUI, and the old `web` server implementation has been removed.

### Development

The `dev.sh` script provides a convenient way to build and run the application for development purposes:

```sh
./dev.sh
```

This script builds the application and then launches the TUI (the `report chores` invocation winds up starting the interface for now).

## Development Conventions

### Code Style

The code follows standard Go formatting and conventions.

### Testing

The project includes tests in the `data` and `util` directories (e.g., `data_test.go`, `status_test.go`, `util_test.go`). To run the tests, use the standard Go test command:

```sh
go test ./...
```

### Dependencies

Project dependencies are managed using Go modules and are defined in the `go.mod` file.
