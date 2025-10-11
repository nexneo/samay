package data

import (
	"context"
	"testing"
	"time"
)

func TestEntryNilGetters(t *testing.T) {
	var entry *Entry

	if entry.GetContent() != "" {
		t.Fatalf("expected empty content for nil entry")
	}
	if entry.GetDuration() != 0 {
		t.Fatalf("expected zero duration for nil entry")
	}
	if entry.GetBillable() {
		t.Fatalf("expected nil entry to be non-billable")
	}
	if entry.GetTags() != nil {
		t.Fatalf("expected nil entry tags to be nil")
	}
	started, err := entry.StartedTime()
	if err != nil {
		t.Fatalf("unexpected error fetching started time: %v", err)
	}
	if started != nil {
		t.Fatalf("expected nil started time for nil entry")
	}
	ended, err := entry.EndedTime()
	if err != nil {
		t.Fatalf("unexpected error fetching ended time: %v", err)
	}
	if ended != nil {
		t.Fatalf("expected nil ended time for nil entry")
	}
}

func TestEntrySaveUpdateDelete(t *testing.T) {
	db := openTempDatabase(t)

	project, err := db.CreateProject("Entry Project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	ctx := context.Background()

	start := time.Now().Add(-10 * time.Minute).UTC()
	end := start.Add(10 * time.Minute)

	entry := &Entry{
		db:         db,
		Project:    project,
		Content:    "Initial work #One #Two ",
		DurationMs: int64((10 * time.Minute) / time.Millisecond),
		StartedAt:  &start,
		EndedAt:    &end,
		Type:       EntryTypeWork,
		Billable:   true,
		Tags:       []string{"One", " two", "ONE"},
	}

	if err := entry.Save(ctx); err != nil {
		t.Fatalf("save entry: %v", err)
	}
	if entry.ID == "" {
		t.Fatalf("expected entry ID to be set after save")
	}
	if entry.ProjectID != project.ID {
		t.Fatalf("expected project ID %d, got %d", project.ID, entry.ProjectID)
	}
	if entry.CreatedAt.IsZero() || entry.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps to be set after save")
	}
	if len(entry.Tags) != 2 || entry.Tags[0] != "One" || entry.Tags[1] != "two" {
		t.Fatalf("expected normalized tags, got %v", entry.Tags)
	}

	entry.Content = "Updated entry #Focus"
	entry.Tags = []string{"Focus", " focus ", "Secondary"}
	entry.DurationMs = int64((15 * time.Minute) / time.Millisecond)
	entry.Billable = false
	entry.Type = EntryTypeChore
	newEnd := time.Now().UTC()
	entry.EndedAt = &newEnd

	if err := entry.Update(ctx); err != nil {
		t.Fatalf("update entry: %v", err)
	}
	if entry.CreatedAt.After(entry.UpdatedAt) {
		t.Fatalf("expected updated timestamp to be >= created timestamp")
	}
	if entry.Billable {
		t.Fatalf("expected entry to be non-billable after update")
	}
	if entry.Type != EntryTypeChore {
		t.Fatalf("expected entry type to be updated")
	}
	if len(entry.Tags) != 2 || entry.Tags[0] != "Focus" || entry.Tags[1] != "Secondary" {
		t.Fatalf("expected updated normalized tags, got %v", entry.Tags)
	}

	storedEntries := project.Entries()
	if storedEntries == nil {
		t.Fatalf("expected entries slice, got nil")
	}
	if len(storedEntries) != 1 {
		t.Fatalf("expected 1 stored entry, got %d", len(storedEntries))
	}
	if storedEntries[0].Content != entry.Content {
		t.Fatalf("expected stored content to match update")
	}
	if storedEntries[0].Billable {
		t.Fatalf("expected stored entry to be non-billable")
	}

	if err := entry.Delete(ctx); err != nil {
		t.Fatalf("delete entry: %v", err)
	}

	postDeleteEntries := project.Entries()
	if len(postDeleteEntries) != 0 {
		t.Fatalf("expected no entries after delete, got %d", len(postDeleteEntries))
	}
}

func TestEntryMinutesAndAccessors(t *testing.T) {
	start := time.Now().Add(-90 * time.Minute).UTC()
	end := start.Add(45 * time.Minute)
	entry := &Entry{
		Content:    "content",
		DurationMs: int64((45 * time.Minute) / time.Millisecond),
		StartedAt:  &start,
		EndedAt:    &end,
		Billable:   true,
		Tags:       []string{"TagOne"},
	}

	if entry.Minutes() != 45 {
		t.Fatalf("expected 45 minutes, got %v", entry.Minutes())
	}
	hm := entry.HoursMins()
	if hm.Hours != 0 || hm.Mins != 45 {
		t.Fatalf("expected 0h45m, got %dh%dm", hm.Hours, hm.Mins)
	}
	if !entry.GetBillable() {
		t.Fatalf("expected entry to be billable")
	}
	if got := entry.GetContent(); got != "content" {
		t.Fatalf("expected content getter to match, got %q", got)
	}
	if entry.GetDuration() != (45 * time.Minute).Nanoseconds() {
		t.Fatalf("expected duration getter to return nanoseconds")
	}
	tags := entry.GetTags()
	if len(tags) != 1 || tags[0] != "TagOne" {
		t.Fatalf("expected copy of tags slice, got %v", tags)
	}
	tags[0] = "mutated"
	if entry.Tags[0] != "TagOne" {
		t.Fatalf("expected GetTags to return copy, entry tags mutated: %v", entry.Tags)
	}
	started, err := entry.StartedTime()
	if err != nil {
		t.Fatalf("started time error: %v", err)
	}
	if started == nil || !started.Equal(start) {
		t.Fatalf("expected started time pointer, got %v", started)
	}
	ended, err := entry.EndedTime()
	if err != nil {
		t.Fatalf("ended time error: %v", err)
	}
	if ended == nil || !ended.Equal(end) {
		t.Fatalf("expected ended time pointer, got %v", ended)
	}
}

func TestEntryMoveToProject(t *testing.T) {
	db := openTempDatabase(t)

	source, err := db.CreateProject("Source")
	if err != nil {
		t.Fatalf("create source project: %v", err)
	}
	target, err := db.CreateProject("Target")
	if err != nil {
		t.Fatalf("create target project: %v", err)
	}

	entry := &Entry{
		db:        db,
		Project:   source,
		Content:   "move me",
		Type:      EntryTypeWork,
		Billable:  false,
		Tags:      []string{"Before"},
		StartedAt: nil,
	}
	if err := entry.SaveNow(); err != nil {
		t.Fatalf("save entry: %v", err)
	}

	if err := entry.MoveToProject(target); err != nil {
		t.Fatalf("move to project: %v", err)
	}
	if entry.ProjectID != target.ID {
		t.Fatalf("expected project ID to be %d, got %d", target.ID, entry.ProjectID)
	}
	if entry.Project.Name != "Target" {
		t.Fatalf("expected project pointer to update")
	}

	for _, stored := range target.Entries() {
		if stored.ID == entry.ID && stored.ProjectID == target.ID {
			return
		}
	}
	t.Fatalf("expected entry to be visible under target project")
}
