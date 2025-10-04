Samay
=====

Command-line Time tracking and reporting
----------------------------------------

Samay is a command-line time tracking tool for developers. It's designed for those who are comfortable working in the terminal. The name "Samay" is the Hindi word for "Time."

The tool allows you to start and stop timers for different projects, manually log time entries, and generate reports. It also includes a web interface to visualize project data and a terminal user interface (TUI) for interactive use.

Why?
----

*   I never find a time tracker that I like to use.
*   My terminal is always open when my laptop is open.
*   I wanted to learn Go and Protocol Buffers.

So, here it is.

Unique features
---------------

*   Simple command-line interface.
*   Interactive Terminal User Interface (TUI).
*   Uses simple files to store data.
*   Can detect and use a Dropbox folder for data synchronization.
*   Reasonably fast.
*   Basic monthly reporting.
*   Web interface for data visualization.

Getting Started
---------------

### Building from source

To build the application from the source code, you'll need to have Go installed.

1.  **Clone the repository:**

    ```sh
    git clone https://github.com/nexneo/samay.git
    cd samay
    ```

2.  **Build the application:**

    ```sh
    go build
    ```

    This will create an executable file named `samay` in the project's root directory.

Usage
-----

The application is run from the command line with various commands and flags.

### Interactive TUI

To use the interactive TUI, run `samay` without any arguments:

```sh
./samay
```

This will open a terminal user interface where you can manage your projects and time entries.

### Command-Line Interface

#### Start/Stop timer

```sh
# Start the timer for a project
./samay start "My Project"

# Stop the timer and add a message
./samay stop "My Project" -m "Finished the first feature"
```

If you don't specify a message with `-m`, your default editor will open to enter the message.

#### Log time directly

```sh
# Log 1.5 hours for a project
./samay entry "My Project" -d 1.5h -m "Team meeting"
```

A duration string is a sequence of decimal numbers, each with an optional fraction and a unit suffix, such as "300m", "1.5h", or "2h45m". Valid time units are "s", "m", and "h".

Twitter-style #hashtags are supported in the log message. For example, a message like "Some time spent in #project #management" will create two tags, "project" and "management", for that entry.

#### Reporting

```sh
# Generate a report for the current month
./samay report

# Generate a report for a specific month (e.g., March)
./samay report -r 3
```

#### Web interface

```sh
# Start the web interface
./samay web
```

This will start a web server on port 8080 by default. You can access it at `http://localhost:8080`.

#### Other commands

*   `./samay log "My Project"`: Show the last 10 time entries for a project.
*   `./samay remove "My Project"`: Remove a project and all its data.

Development
-----------

### Running tests

To run the tests, use the standard Go test command:

```sh
go test ./...
```

### Dependencies

Project dependencies are managed using Go modules and are defined in the `go.mod` file. The main dependencies are:

*   **Protocol Buffers:** For data serialization.
*   **Bubble Tea:** For the terminal user interface.
*   **Gorilla Mux:** For the web server.

Caveats
-------

*   If you have a folder named "Samay" in your Dropbox root, please rename it before running this utility.
*   This software comes with NO WARRANTIES.