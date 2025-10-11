package data

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func setupTestDatabase(t *testing.T) *Database {
	t.Helper()

	path := filepath.Join(".", "test.db")
	// Ensure leftover files from a previous run do not interfere with this test.
	cleanupSQLiteArtifacts(t, path)

	db, err := open(path)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
		cleanupSQLiteArtifacts(t, path)
	})

	return db
}

func cleanupSQLiteArtifacts(t *testing.T, path string) {
	t.Helper()

	remove := func(target string) {
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("remove %s: %v", target, err)
		}
	}

	remove(path)
	remove(path + "-wal")
	remove(path + "-shm")
}

func TestDatabaseProjectLifecycle(t *testing.T) {
	db := setupTestDatabase(t)

	project, err := db.CreateProject("Test Project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.ID == 0 {
		t.Fatalf("expected project ID to be set")
	}
	if project.Name != "Test Project" {
		t.Fatalf("expected project name to be %q, got %q", "Test Project", project.Name)
	}

	if _, err := db.CreateProject("Test Project"); err == nil {
		t.Fatalf("expected duplicate project creation to fail")
	}

	projects := db.Projects()
	if len(projects) != 1 {
		t.Fatalf("expected exactly one project, got %d", len(projects))
	}
	if projects[0].Name != "Test Project" {
		t.Fatalf("expected project list name to be %q, got %q", "Test Project", projects[0].Name)
	}

	if err := project.Rename("Renamed Project"); err != nil {
		t.Fatalf("rename project: %v", err)
	}
	if project.Name != "Renamed Project" {
		t.Fatalf("expected renamed project to be %q, got %q", "Renamed Project", project.Name)
	}

	if err := project.Rename("   "); err == nil {
		t.Fatalf("expected rename to empty string to fail")
	}

	if err := project.Delete(); err != nil {
		t.Fatalf("delete project: %v", err)
	}

	if remaining := db.Projects(); len(remaining) != 0 {
		t.Fatalf("expected projects to be empty after delete, got %d", len(remaining))
	}
}
