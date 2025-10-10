# Repository Guidelines

## Project Structure & Module Organization
Samay is a single Go module (`go.mod`) with the entry point in `main.go`. Core persistence logic lives in `data/`, which owns the SQLite schema (`sql/schema.sql`), `sqlc`-generated code, and helpers for working with the database. The terminal interface is kept in `tui/`, while shared helpers sit in `util/`. HTML assets for the now-deprecated web mode remain in `public/` for future reuse. Tests are colocated with their packages (for example `data/..._test.go` and `util/util_test.go`).

## Build, Test, and Development Commands
Use `go build` from the repo root to compile the `samay` binary. Run `./dev.sh` to rebuild and launch the Bubble Tea TUI in one step during development. Execute `go test ./...` to run the full suite, and add `-cover` when you need coverage numbers.

## Coding Style & Naming Conventions
Stick to idiomatic Go style: tabs for indentation, `gofmt` and `goimports` keep code formatted. Exported identifiers follow `CamelCase`; tests and private helpers use `lowerCamelCase`. When the SQLite schema or queries change (`data/sql/*.sql`), rerun `sqlc generate` so the typed data access layer stays in sync. Keep file names snake_case in `tui/` and mirror existing patterns when adding new screens.

When deferring cleanup (for example, closing files or database handles), make sure to check and propagate errors returned by the deferred call so `errcheck` and similar linters stay satisfied.

## Testing Guidelines
Write tests alongside the code under test using Go's `testing` package. Name suites with `*_test.go` and individual cases with `TestScenarioName`. Exercise persistence flows through the SQLite-backed data layer and state transitions in `tui/`. Before opening a pull request, run `go test ./...` (or `go test -run <Package> ./...` for focused work) to ensure a clean signal.

## Commit & Pull Request Guidelines
Commits in this repo use short imperative subjects; conventional prefixes like `feat:` or `fix:` are welcome but not required. Keep changes focused and include context in the body when behavior shifts. Pull requests should summarize the rationale, reference relevant issues, and note user-facing impacts. Screenshots or terminal captures help reviewers when TUI behavior changes. Always mention required follow-up tasks, such as rerunning `sqlc generate`, in the PR description.
