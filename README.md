# Samay

Terminal-first time tracking and reporting.

Samay is a Go application that lets you track work entirely from the terminal. It stores data in a local SQLite database, offers a Bubble Tea interface for daily use, and keeps the old command-line workflows around as future enhancements.

<p align="center">
  <img alt="Weekly overview" src="https://github.com/user-attachments/assets/48605f74-69d6-4eee-a65f-b1b29253ee45" />
</p>

## Why?

- Frustration with existing time trackers, slow, complex.
- The terminal is always open.
- A playground for learning Go, SQL, and Bubble Tea.

## Highlights

- TUI for starting and stopping timers, adding manual entries, and reviewing history.
- Data stored in a single SQLite database (defaults to `~/Documents/Samay.db`, configurable on first launch).
- Built-in monthly report and weekly overview views.

## Prerequisites

- Go 1.23+ (the repo is configured with the Go 1.24.2 toolchain).
- SQLite is bundled; Samay uses the pure Go `modernc.org/sqlite` driver, so no external system packages are required.

## Installation

Install Samay with the Go toolchain so you always pull the latest tagged release. On macOS the shortest path is to use Homebrew for Go itself, then install Samay in one step:

```sh
brew install go
go install github.com/nexneo/samay@latest
```

The binary ends up at `$(go env GOPATH)/bin/samay` (typically `~/go/bin/samay`); add that directory to your `PATH` if it is not already there.

## Version

Check the installed version with:

```sh
samay --version
```

When cutting a release, update the constant in `util/version/version.go` and create a matching Git tag (for example `v1.0.0`) so `go install github.com/nexneo/samay@latest` resolves to the correct binary.

## Running Samay

Launch the interface with:

```sh
./samay
```

On the first launch Samay asks where to create the SQLite database (press Enter to accept the default `~/Documents/Samay.db`). Subsequent runs reuse the saved location.

The TUI opens to a project list backed by that database. Use the arrow keys (or `j`/`k`) to highlight a project and from there:

- `s` starts a timer. The project is persisted as soon as you start tracking against it.
- `p` stops the active timer and prompts for a summary message.
- `e` records a manual entry—enter a duration such as `45m` or `1h30m`, then the description.
- `l` shows the project log with scrollable history (`↑/↓/PgUp/PgDn`), and `a` toggles between recent entries and the full timeline.
- `v` lists entries so you can review details, move them to another project, or delete them.
- `r` renames the project; `d` deletes it.

At the project list level, press `r` to open the monthly report for the highlighted month and `o` for the weekly overview dashboard. `Esc` navigates back; `q` quits from anywhere.

## Data Storage

Samay persists everything in a single SQLite database. The default location is `~/Documents/Samay.db`, but you can point it anywhere on disk. The schema tracks:

- `projects`: project metadata plus timestamps and a hidden flag.
- `entries`: individual time entries with nanosecond precision duration, start/stop timestamps, billable flag, and optional creator.
- `entry_tags`: many-to-many join table for hashtag extraction.
- `timers`: one active timer per project.

The schema lives in `data/sql/schema.sql` and the sqlc query definitions are in `data/sql/queries.sql`.

## Development & Testing

Run the full test suite with:

```sh
go test ./...
```

If you change `data/sql/schema.sql` or `data/sql/queries.sql`, regenerate the typed data access layer with:

```sh
sqlc generate
```

Dependencies are managed through Go modules; see `go.mod` for the current set.

## Troubleshooting

- The database location is stored in `$(os.UserConfigDir())/samay/config.json`. Delete that file to re-run the first-time setup prompt.
- Because the SQLite driver is pure Go, the binary runs anywhere Go runs—no CGO or native SQLite is required.
- Samay is provided without warranty—use at your own risk.
