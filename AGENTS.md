# Gemini Code Understanding

## Project Overview

This project, "Samay," is a command-line time tracking tool written in Go. It's designed for developers who are comfortable working in the terminal. The name "Samay" is the Hindi word for "Time."

The tool allows users to start and stop timers for different projects, manually log time entries, and generate reports. It also includes a web interface to visualize project data. Data is stored locally in files, and the application can detect and use a Dropbox folder for data synchronization.

The project uses Google's Protocol Buffers for data serialization, the `bubbletea` library for its terminal user interface (TUI), and the `gorilla/mux` router for its web component.

## Building and Running

### Building the application

To build the application, run the following command:

```sh
go build
```

This will create an executable file named `samay` in the project's root directory.

### Running the application

The application is run from the command line with various commands and flags.

**Starting and stopping the timer:**

```sh
# Start the timer for a project
./samay start "My Project"

# Stop the timer and add a message
./samay stop "My Project" -m "Finished the first feature"
```

**Logging time directly:**

```sh
# Log 1.5 hours for a project
./samay entry "My Project" -d 1.5h -m "Team meeting"
```

**Generating reports:**

```sh
# Generate a report for the current month
./samay report

# Generate a report for a specific month (e.g., March)
./samay report -r 3
```

**Running the web interface:**

```sh
./samay web
```

This will start a web server on port 8080 by default. You can access it at `http://localhost:8080`.

### Development

The `dev.sh` script provides a convenient way to build and run the application for development purposes:

```sh
./dev.sh
```

This script builds the application and then runs the `report chores` command.

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
