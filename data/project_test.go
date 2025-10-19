package data

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nexneo/samay/data/sqlc"
)

func TestProjectGettersNil(t *testing.T) {
	var project *Project

	if project.GetName() != "" {
		t.Fatalf("expected empty name for nil project")
	}
	if project.GetCompany() != "" {
		t.Fatalf("expected empty company for nil project")
	}
	if entries := project.Entries(); entries != nil {
		t.Fatalf("expected nil entries slice for nil project")
	}
	if err := project.StartTimer(); err == nil {
		t.Fatalf("expected start timer on nil project to error")
	}
	if err := project.StopTimer("content", false); err == nil {
		t.Fatalf("expected stop timer on nil project to error")
	}
}

func TestProjectEntriesAndTimerFlow(t *testing.T) {
	db := openTempDatabase(t)

	project, err := db.CreateProject("Timer Project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if err := project.StartTimer(); err != nil {
		t.Fatalf("start timer: %v", err)
	}

	onClock, timer := project.OnClock()
	if !onClock || timer == nil {
		t.Fatalf("expected timer to be running")
	}

	ctx := context.Background()
	customStart := time.Now().Add(-90 * time.Second).UTC().Truncate(time.Second)
	if _, err := db.queries.UpsertTimer(ctx, sqlc.UpsertTimerParams{
		ProjectID: project.ID,
		StartedAt: customStart.Unix(),
	}); err != nil {
		t.Fatalf("backdate timer start: %v", err)
	}

	content := "  Working on timers #Focus "
	if err := project.StopTimer(content, true); err != nil {
		t.Fatalf("stop timer: %v", err)
	}

	onClock, _ = project.OnClock()
	if onClock {
		t.Fatalf("expected timer to be cleared after stop")
	}

	entries := project.Entries()
	if entries == nil {
		t.Fatalf("expected entries slice, got nil")
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Content != "Working on timers #Focus" {
		t.Fatalf("unexpected entry content: %q", entry.Content)
	}
	if !entry.Billable {
		t.Fatalf("expected entry to be billable")
	}
	if len(entry.Tags) != 1 || entry.Tags[0] != "Focus" {
		t.Fatalf("expected entry tags to contain Focus, got %v", entry.Tags)
	}
	if entry.DurationMs < 80_000 || entry.DurationMs > 100_000 {
		t.Fatalf("expected entry duration close to 90s, got %d", entry.DurationMs)
	}
}

func TestProjectStopTimerWithoutActive(t *testing.T) {
	db := openTempDatabase(t)

	project, err := db.CreateProject("No Timer")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	err = project.StopTimer("content", false)
	if err == nil {
		t.Fatalf("expected error when stopping timer without active timer")
	}
	if err.Error() != "no running timer for project" {
		t.Fatalf("unexpected error stopping timer: %v", err)
	}
}

func TestProjectCreateEntryVariants(t *testing.T) {
	db := openTempDatabase(t)

	project, err := db.CreateProject("Entries Project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	entry, err := project.CreateEntry("  Writing docs #Docs  ", true)
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if entry.ProjectID != project.ID {
		t.Fatalf("expected project ID %d, got %d", project.ID, entry.ProjectID)
	}
	if entry.Content != "Writing docs #Docs" {
		t.Fatalf("unexpected content: %q", entry.Content)
	}
	if entry.DurationMs != 0 {
		t.Fatalf("expected zero duration for quick entry, got %d", entry.DurationMs)
	}
	if entry.StartedAt == nil || entry.StartedAt.IsZero() {
		t.Fatalf("expected StartedAt to be set")
	}
	if len(entry.Tags) != 1 || entry.Tags[0] != "Docs" {
		t.Fatalf("expected Docs tag, got %v", entry.Tags)
	}

	duration := 45 * time.Second
	entryWithDuration, err := project.CreateEntryWithDuration("Meeting review #Review", duration, false)
	if err != nil {
		t.Fatalf("create entry with duration: %v", err)
	}
	if entryWithDuration.ProjectID != project.ID {
		t.Fatalf("expected project ID %d, got %d", project.ID, entryWithDuration.ProjectID)
	}
	if entryWithDuration.DurationMs != duration.Milliseconds() {
		t.Fatalf("expected duration %dms, got %d", duration.Milliseconds(), entryWithDuration.DurationMs)
	}
	if entryWithDuration.StartedAt == nil || entryWithDuration.EndedAt == nil {
		t.Fatalf("expected Start/End to be set")
	}
	if entryWithDuration.EndedAt.Sub(*entryWithDuration.StartedAt) != duration {
		t.Fatalf("expected duration between timestamps to equal %v", duration)
	}
	if entryWithDuration.Billable {
		t.Fatalf("expected entry to be non-billable")
	}
	if len(entryWithDuration.Tags) != 1 || entryWithDuration.Tags[0] != "Review" {
		t.Fatalf("expected Review tag, got %v", entryWithDuration.Tags)
	}
}

func TestProjectsSortedByPositionAndUpdatedAt(t *testing.T) {
	db := openTempDatabase(t)

	alpha, err := db.CreateProject("Alpha")
	if err != nil {
		t.Fatalf("create alpha project: %v", err)
	}
	bravo, err := db.CreateProject("Bravo")
	if err != nil {
		t.Fatalf("create bravo project: %v", err)
	}
	charlie, err := db.CreateProject("Charlie")
	if err != nil {
		t.Fatalf("create charlie project: %v", err)
	}

	ctx := context.Background()
	now := time.Now().UTC().Unix()

	updates := []struct {
		id       int64
		position int64
		updated  int64
	}{
		{alpha.ID, 2, now},
		{bravo.ID, 1, now + 1},
		{charlie.ID, 1, now + 2},
	}

	for _, u := range updates {
		if _, err := db.sqlite.ExecContext(ctx, "UPDATE projects SET position = ?, updated_at = ? WHERE id = ?", u.position, u.updated, u.id); err != nil {
			t.Fatalf("update project %d: %v", u.id, err)
		}
	}

	projects := db.Projects()
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(projects))
	}

	got := []string{projects[0].Name, projects[1].Name, projects[2].Name}
	want := []string{"Charlie", "Bravo", "Alpha"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected order at index %d: got %q, want %q (full order: %v)", i, got[i], want[i], got)
		}
	}

	if projects[0].Position != 1 || projects[1].Position != 1 || projects[2].Position != 2 {
		t.Fatalf("unexpected project positions: %v", []int64{projects[0].Position, projects[1].Position, projects[2].Position})
	}
}

func TestProjectRenameNoOp(t *testing.T) {
	db := openTempDatabase(t)

	project, err := db.CreateProject("Rename Me")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	updatedAt := project.UpdatedAt

	if err := project.Rename("rename me"); err != nil {
		t.Fatalf("rename case-insensitive: %v", err)
	}
	if project.Name != "Rename Me" {
		t.Fatalf("expected name to remain unchanged, got %q", project.Name)
	}
	if !project.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected timestamps to remain unchanged on no-op rename")
	}

	if err := project.Rename("  New Name   "); err != nil {
		t.Fatalf("rename trimmed: %v", err)
	}
	if project.Name != "New Name" {
		t.Fatalf("expected trimmed new name, got %q", project.Name)
	}
	if strings.TrimSpace(project.Name) != project.Name {
		t.Fatalf("expected name to be stored trimmed")
	}
}
