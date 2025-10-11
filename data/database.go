package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nexneo/samay/data/sqlc"
	_ "modernc.org/sqlite"
)

var (
	// DB is the global database handle configured during program start.
	DB *Database

	openOnce sync.Once
)

// Database wraps the SQLite connection, precompiled sqlc queries, and metadata.
type Database struct {
	path    string
	sqlite  *sql.DB
	queries *sqlc.Queries
}

// OpenDatabase initializes the global DB handle using the provided SQLite path.
// It is safe to call multiple times; the first successful call wins.
func OpenDatabase(path string) (err error) {
	openOnce.Do(func() {
		var db *Database
		db, err = open(path)
		DB = db
	})
	return err
}

// open constructs a Database instance, configuring pragmas and ensuring schema.
func open(path string) (*Database, error) {
	if path == "" {
		return nil, errors.New("database path cannot be empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o775); err != nil {
		return nil, fmt.Errorf("prepare database directory: %w", err)
	}

	sqlite, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := configureSQLite(sqlite); err != nil {
		if closeErr := sqlite.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close sqlite: %w", closeErr))
		}
		return nil, err
	}

	db := &Database{
		path:    path,
		sqlite:  sqlite,
		queries: sqlc.New(sqlite),
	}

	if err := db.ensureSchema(context.Background()); err != nil {
		if closeErr := sqlite.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close sqlite: %w", closeErr))
		}
		return nil, err
	}

	return db, nil
}

// Close shuts down the underlying SQLite connection.
func (d *Database) Close() error {
	if d == nil || d.sqlite == nil {
		return nil
	}
	return d.sqlite.Close()
}

// Path exposes the on-disk location backing the database.
func (d *Database) Path() string {
	if d == nil {
		return ""
	}
	return d.path
}

// Queries returns the sqlc-generated query helper bound to this database.
func (d *Database) Queries() *sqlc.Queries {
	if d == nil {
		return nil
	}
	return d.queries
}

func configureSQLite(db *sql.DB) error {
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
	}

	for _, stmt := range pragmas {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("set sqlite pragma %q: %w", stmt, err)
		}
	}
	return nil
}

func (d *Database) ensureSchema(ctx context.Context) error {
	statements := splitStatements(schemaSQL)
	for _, stmt := range statements {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := d.sqlite.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("apply schema statement %q: %w", stmt, err)
		}
	}
	return nil
}

func splitStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			statements = append(statements, trimmed)
		}
	}
	return statements
}
