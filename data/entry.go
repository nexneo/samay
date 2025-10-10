package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nexneo/samay/data/sqlc"
	"github.com/nexneo/samay/util"
)

type EntryType string

const (
	EntryTypeChore EntryType = "CHORE"
	EntryTypeFun   EntryType = "FUN"
	EntryTypeWork  EntryType = "WORK"
)

var tagFinder = regexp.MustCompile(`\B#(\w\w+)`)

type Entry struct {
	db         *Database
	Project    *Project
	ID         string
	ProjectID  int64
	CreatorID  *int64
	Content    string
	DurationMs int64
	StartedAt  *time.Time
	EndedAt    *time.Time
	Type       EntryType
	Billable   bool
	Tags       []string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func newEntryFromModel(db *Database, project *Project, model sqlc.Entry) *Entry {
	var started *time.Time
	if model.StartedAt.Valid {
		t := time.Unix(model.StartedAt.Int64, 0).UTC()
		started = &t
	}
	var ended *time.Time
	if model.EndedAt.Valid {
		t := time.Unix(model.EndedAt.Int64, 0).UTC()
		ended = &t
	}
	var creatorID *int64
	if model.CreatorID.Valid {
		id := model.CreatorID.Int64
		creatorID = &id
	}
	return &Entry{
		db:         db,
		Project:    project,
		ID:         model.ID,
		ProjectID:  model.ProjectID,
		CreatorID:  creatorID,
		Content:    model.Content,
		DurationMs: model.DurationMs,
		StartedAt:  started,
		EndedAt:    ended,
		Type:       EntryType(model.EntryType),
		Billable:   model.IsBillable == 1,
		CreatedAt:  time.Unix(model.CreatedAt, 0).UTC(),
		UpdatedAt:  time.Unix(model.UpdatedAt, 0).UTC(),
	}
}

func (e *Entry) GetContent() string {
	if e == nil {
		return ""
	}
	return e.Content
}

func (e *Entry) GetDuration() int64 {
	if e == nil {
		return 0
	}
	return e.DurationMs * int64(time.Millisecond)
}

func (e *Entry) GetBillable() bool {
	if e == nil {
		return false
	}
	return e.Billable
}

func (e *Entry) GetTags() []string {
	if e == nil {
		return nil
	}
	return append([]string(nil), e.Tags...)
}

func (e *Entry) StartedTime() (*time.Time, error) {
	if e == nil || e.StartedAt == nil {
		return nil, nil
	}
	return e.StartedAt, nil
}

func (e *Entry) EndedTime() (*time.Time, error) {
	if e == nil || e.EndedAt == nil {
		return nil, nil
	}
	return e.EndedAt, nil
}

func (e *Entry) Minutes() float64 {
	return time.Duration(e.GetDuration()).Minutes()
}

func (e *Entry) HoursMins() hoursMins {
	return hoursMinsFromDuration(time.Duration(e.GetDuration()))
}

func (e *Entry) Save(ctx context.Context) error {
	if e == nil {
		return errors.New("entry is nil")
	}
	if e.db == nil {
		return errors.New("entry has no database reference")
	}
	if e.Project == nil {
		return errors.New("entry has no project reference")
	}
	if e.ProjectID == 0 {
		e.ProjectID = e.Project.ID
	}
	if e.ProjectID == 0 {
		return errors.New("entry missing project id")
	}

	if e.ID == "" {
		if id, err := util.UUID(); err == nil {
			e.ID = id
		} else {
			return fmt.Errorf("generate entry id: %w", err)
		}
	}

	var creator sql.NullInt64
	if e.CreatorID != nil {
		creator = sql.NullInt64{Int64: *e.CreatorID, Valid: true}
	}
	var started sql.NullInt64
	if e.StartedAt != nil {
		started = sql.NullInt64{Int64: e.StartedAt.Unix(), Valid: true}
	}
	var ended sql.NullInt64
	if e.EndedAt != nil {
		ended = sql.NullInt64{Int64: e.EndedAt.Unix(), Valid: true}
	}

	params := sqlc.CreateEntryParams{
		ID:         e.ID,
		ProjectID:  e.ProjectID,
		CreatorID:  creator,
		Content:    e.Content,
		DurationMs: e.DurationMs,
		StartedAt:  started,
		EndedAt:    ended,
		EntryType:  string(e.Type),
		IsBillable: boolToInt(e.Billable),
	}

	record, err := e.db.queries.CreateEntry(ctx, params)
	if err != nil {
		return fmt.Errorf("insert entry: %w", err)
	}
	e.CreatedAt = time.Unix(record.CreatedAt, 0).UTC()
	e.UpdatedAt = time.Unix(record.UpdatedAt, 0).UTC()

	if err := e.replaceTags(ctx); err != nil {
		return err
	}
	return nil
}

func (e *Entry) Update(ctx context.Context) error {
	if e == nil {
		return errors.New("entry is nil")
	}
	if e.db == nil {
		return errors.New("entry has no database reference")
	}
	if e.Project != nil {
		e.ProjectID = e.Project.ID
	}
	if e.ProjectID == 0 {
		return errors.New("entry missing project id")
	}
	if e.ID == "" {
		return errors.New("entry missing identifier")
	}

	var creator sql.NullInt64
	if e.CreatorID != nil {
		creator = sql.NullInt64{Int64: *e.CreatorID, Valid: true}
	}
	var started sql.NullInt64
	if e.StartedAt != nil {
		started = sql.NullInt64{Int64: e.StartedAt.Unix(), Valid: true}
	}
	var ended sql.NullInt64
	if e.EndedAt != nil {
		ended = sql.NullInt64{Int64: e.EndedAt.Unix(), Valid: true}
	}

	record, err := e.db.queries.UpdateEntry(ctx, sqlc.UpdateEntryParams{
		ID:         e.ID,
		ProjectID:  e.ProjectID,
		CreatorID:  creator,
		Content:    e.Content,
		DurationMs: e.DurationMs,
		StartedAt:  started,
		EndedAt:    ended,
		EntryType:  string(e.Type),
		IsBillable: boolToInt(e.Billable),
	})
	if err != nil {
		return fmt.Errorf("update entry: %w", err)
	}
	e.CreatedAt = time.Unix(record.CreatedAt, 0).UTC()
	e.UpdatedAt = time.Unix(record.UpdatedAt, 0).UTC()
	if err := e.replaceTags(ctx); err != nil {
		return err
	}
	return nil
}

func (e *Entry) Delete(ctx context.Context) error {
	if e == nil || e.db == nil {
		return errors.New("entry not initialized")
	}
	if e.ID == "" {
		return errors.New("entry missing identifier")
	}
	if err := e.db.queries.DeleteEntry(ctx, e.ID); err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}
	return nil
}

func (e *Entry) MoveTo(ctx context.Context, project *Project) error {
	if project == nil {
		return errors.New("target project is nil")
	}
	e.Project = project
	e.ProjectID = project.ID
	return e.Update(ctx)
}

func (e *Entry) SaveNow() error {
	return e.Save(context.Background())
}

func (e *Entry) UpdateNow() error {
	return e.Update(context.Background())
}

func (e *Entry) DeleteNow() error {
	return e.Delete(context.Background())
}

func (e *Entry) MoveToProject(project *Project) error {
	return e.MoveTo(context.Background(), project)
}

func (e *Entry) replaceTags(ctx context.Context) error {
	if err := e.db.queries.DeleteEntryTags(ctx, e.ID); err != nil {
		return fmt.Errorf("clear entry tags: %w", err)
	}
	normalized := uniqueSortedTags(e.Tags)
	e.Tags = normalized
	for _, tag := range normalized {
		if err := e.db.queries.InsertEntryTag(ctx, sqlc.InsertEntryTagParams{EntryID: e.ID, Tag: tag}); err != nil {
			return fmt.Errorf("insert tag %q: %w", tag, err)
		}
	}
	return nil
}

func uniqueSortedTags(tags []string) []string {
	dedup := make(map[string]string, len(tags))
	keys := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := dedup[key]; !exists {
			dedup[key] = trimmed
			keys = append(keys, key)
		}
	}
	result := make([]string, 0, len(dedup))
	for _, key := range keys {
		result = append(result, dedup[key])
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i]) < strings.ToLower(result[j])
	})
	return result
}

func extractTags(content string) []string {
	matches := tagFinder.FindAllStringSubmatch(content, 20)
	tags := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, match[1])
		}
	}
	return uniqueSortedTags(tags)
}
