package data

import (
	"path/filepath"
	"testing"
)

func openTempDatabase(t *testing.T) *Database {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := open(path)
	if err != nil {
		t.Fatalf("open temp database: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close temp database: %v", err)
		}
		cleanupSQLiteArtifacts(t, path)
	})

	return db
}
