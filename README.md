# Samay

Terminal-first time tracking and reporting.

Samay is a Go application that lets you track work entirely from the terminal. It stores data locally inside your Dropbox folder, offers a Bubble Tea interface for daily use, and keeps the old command-line workflows around as future enhancements.

<p align="center">
  <img alt="Weekly overview" src="https://github.com/user-attachments/assets/8a149da6-4ecd-4c29-9765-b7f365408439" width="55%" />
</p>

## Why?

- Frustration with existing time trackers.
- The terminal is always open.
- A playground for learning Go, Protocol Buffers, and Bubble Tea.

## Highlights

- TUI for starting and stopping timers, adding manual entries, and reviewing history.
- Data stored as Protocol Buffers under `~/Dropbox/Samay` when available, or your local config directory otherwise.
- Built-in monthly report and weekly overview views.

## Prerequisites

- Go 1.23+ (the repo is configured with the Go 1.24.2 toolchain).
- Optional: Dropbox desktop client with a valid `~/.dropbox/host.db` file so Samay can sync via `~/Dropbox/Samay`; otherwise data is stored under your user config directory (for example `~/.config/samay`).
- Make sure there is no existing `~/Dropbox/Samay` directory that you care about before running the app; Samay will create and manage that directory.

## Installation

Install Samay with the Go toolchain so you always pull the latest tagged release. On macOS the shortest path is to use Homebrew for Go itself, then install Samay in one step:

```sh
brew install go
go install github.com/nexneo/samay@latest
```

The binary ends up at `$(go env GOPATH)/bin/samay` (typically `~/go/bin/samay`); add that directory to your `PATH` if it is not already there.

## Running Samay

Launch the interface with:

```sh
./samay
```

The TUI opens to a project list sourced from `~/Dropbox/Samay`. Use the arrow keys (or `j`/`k`) to highlight a project and from there:

- `s` starts a timer. The project is persisted as soon as you start tracking against it.
- `p` stops the active timer and prompts for a summary message.
- `e` records a manual entry—enter a duration such as `45m` or `1h30m`, then the description.
- `l` shows the project log with scrollable history (`↑/↓/PgUp/PgDn`), and `a` toggles between recent entries and the full timeline.
- `v` lists entries so you can review details, move them to another project, or delete them.
- `r` renames the project; `d` deletes it.

At the project list level, press `r` to open the monthly report for the highlighted month and `o` for the weekly overview dashboard. `Esc` navigates back; `q` quits from anywhere.

## Data Storage

Samay keeps each project under a SHA1-named directory inside `~/Dropbox/Samay` (or whichever base path Samay resolves). Every project directory contains a `project.db` file and an `entries/` folder with per-entry protocol buffer records. Timers in progress live next to those entries as `timer.db`. Dropbox synchronizes your tracked time automatically across machines when you use its folder.

## Development & Testing

Run the full test suite with:

```sh
go test ./...
```

Regenerate protocol buffer code after editing `data/models.proto` using `protoc --go_out=. data/models.proto` (requires `protoc` installed). Dependencies are managed through Go modules; see `go.mod` for the current set.

## Troubleshooting

- Set `SAMAY_DATA_DIR` to override the storage directory. When Dropbox metadata is missing, Samay falls back to your user config directory so you can run tests and local builds without Dropbox installed.
- If the interface launches with an empty project list, seed `~/Dropbox/Samay` with at least one project directory (containing a `project.db` file) before restarting. Project creation inside the TUI is on the roadmap.
- Samay is provided without warranty—use at your own risk.
