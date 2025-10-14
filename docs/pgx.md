# Migrating from `database/sql` (SQLite) to `pgx` (PostgreSQL)

This document outlines the high-level work required to replace the current SQLite-backed `database/sql` stack with `pgx` and PostgreSQL.

## 1. Replace the Connection Layer
- Update `data/database.go` to use `pgxpool.Pool` (or another pgx connection type) instead of `*sql.DB`.
- Remove SQLite-only configuration (journal mode, foreign keys, busy timeout) and replace it with PostgreSQL connection tuning.
- Swap imports in `go.mod`/`go.sum`: add `github.com/jackc/pgx/v5`, drop `modernc.org/sqlite`.
- Rework lifecycle helpers (`OpenDatabase`, `Close`, `configureSQLite`) to match pgx behavior.

## 2. Regenerate sqlc for PostgreSQL
- Change `sqlc.yaml` to `engine: postgresql` and configure the Go generator to emit pgx-compatible code (`sql_package: pgx/v5`).
- Rewrite `data/sql/schema.sql` and `data/sql/queries.sql` for PostgreSQL syntax:
  - Replace positional placeholders (`?1`) with `$1`, `$2`, …
  - Drop `PRAGMA` statements and any SQLite-specific clauses (`COLLATE NOCASE`, `WITHOUT ROWID`, `unixepoch()` defaults, integer booleans).
  - Adopt PostgreSQL types (`BOOLEAN`, `TIMESTAMPTZ`, etc.) and default expressions (`NOW()`).
- Run `sqlc generate`; expect the entire `data/sqlc` package to change.

## 3. Update Data Layer Types and Helpers
- Adjust constructors and persistence logic in `data/entry.go`, `data/project.go`, and `data/timer.go` to work with the regenerated pgx structs.
- Replace usages of `sql.Null*` with the pgx equivalents (e.g., `pgtype.Text`, `pgtype.Int8`, plain `*int64` if sqlc emits pointers).
- Remove helper functions that convert booleans to ints (`boolToInt`) and update code paths that depended on SQLite’s integer booleans.

## 4. Handle Schema Initialization and Migrations
- Decide how PostgreSQL schema management should work (manual migrations, `sqlc` bootstrap, or a tool like `goose`/`migrate`).
- Update `data/database.go` so `ensureSchema` runs the appropriate setup or migration logic against PostgreSQL.

## 5. Configuration and Runtime Changes
- Introduce configuration for PostgreSQL connection strings (environment variables or config files).
- Ensure the application can connect to a running PostgreSQL instance during development and testing.

## 6. Testing and Validation
- Update or add tests that assume SQLite behavior (case sensitivity, transaction semantics, timestamps) to reflect PostgreSQL.
- Run `go test ./...` against PostgreSQL-backed test databases.
- Manually validate TUI flows that depend on persistence (project creation, timers, entries).

## 7. Operational Considerations
- Provision PostgreSQL for local development (Docker, Homebrew, etc.) and document setup steps.
- Plan data migration if existing SQLite data must be preserved.
- Update CI/CD pipelines and documentation to require PostgreSQL instead of SQLite.
