package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nexneo/samay/data/sqlc"
)

type Project struct {
	db        *Database
	ID        int64
	Name      string
	Company   *string
	IsHidden  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func newProjectFromModel(db *Database, model sqlc.Project) *Project {
	var company *string
	if model.Company.Valid {
		value := model.Company.String
		company = &value
	}
	return &Project{
		db:        db,
		ID:        model.ID,
		Name:      model.Name,
		Company:   company,
		IsHidden:  model.IsHidden == 1,
		CreatedAt: time.Unix(model.CreatedAt, 0).UTC(),
		UpdatedAt: time.Unix(model.UpdatedAt, 0).UTC(),
	}
}

func (p *Project) GetName() string {
	if p == nil {
		return ""
	}
	return p.Name
}

func (p *Project) GetCompany() string {
	if p == nil || p.Company == nil {
		return ""
	}
	return *p.Company
}

func (p *Project) Entries() []*Entry {
	if p == nil || p.db == nil {
		return nil
	}
	entries, err := p.entries()
	if err != nil {
		fmt.Printf("Failed to list entries for project %s: %v\n", p.Name, err)
		return nil
	}
	return entries
}

func (p *Project) entries() ([]*Entry, error) {
	ctx := context.Background()
	rows, err := p.db.queries.ListEntriesByProject(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	entries := make([]*Entry, 0, len(rows))
	for _, row := range rows {
		e := newEntryFromModel(p.db, p, row)
		tagRows, err := p.db.queries.ListTagsForEntry(ctx, row.ID)
		if err != nil {
			return nil, fmt.Errorf("load tags for entry %s: %w", row.ID, err)
		}
		tags := make([]string, 0, len(tagRows))
		for _, tag := range tagRows {
			tags = append(tags, tag.Tag)
		}
		e.Tags = tags
		entries = append(entries, e)
	}
	return entries, nil
}

func (p *Project) StartTimer() error {
	if p == nil || p.db == nil {
		return errors.New("project not initialized")
	}
	ctx := context.Background()
	_, err := p.db.queries.UpsertTimer(ctx, sqlc.UpsertTimerParams{
		ProjectID: p.ID,
		StartedAt: time.Now().UTC().Unix(),
	})
	if err != nil {
		return fmt.Errorf("start timer: %w", err)
	}
	return nil
}

func (p *Project) StopTimer(content string, billable bool) error {
	if p == nil || p.db == nil {
		return errors.New("project not initialized")
	}

	timer, err := p.currentTimer()
	if err != nil {
		return err
	}
	if timer == nil {
		return errors.New("no running timer for project")
	}

	start := timer.StartedAt
	if start.IsZero() {
		start = time.Now().UTC()
	}
	end := time.Now().UTC()
	if end.Before(start) {
		end = start
	}

	duration := end.Sub(start)
	entry := &Entry{
		db:         p.db,
		Project:    p,
		ProjectID:  p.ID,
		Content:    strings.TrimSpace(content),
		DurationMs: duration.Milliseconds(),
		StartedAt:  &start,
		EndedAt:    &end,
		Type:       EntryTypeWork,
		Billable:   billable,
		Tags:       extractTags(content),
	}

	if err := entry.Save(context.Background()); err != nil {
		return fmt.Errorf("persist timer entry: %w", err)
	}

	if err := p.db.queries.DeleteTimer(context.Background(), p.ID); err != nil {
		return fmt.Errorf("clear timer: %w", err)
	}

	fmt.Printf("%.2f mins\n", entry.Minutes())
	return nil
}

func (p *Project) CreateEntryWithDuration(content string, duration time.Duration, billable bool) (*Entry, error) {
	entry := &Entry{
		db:         p.db,
		Project:    p,
		ProjectID:  p.ID,
		Content:    strings.TrimSpace(content),
		DurationMs: duration.Milliseconds(),
		Type:       EntryTypeWork,
		Billable:   billable,
		Tags:       extractTags(content),
	}
	end := time.Now().UTC()
	start := end.Add(-duration)
	entry.StartedAt = &start
	entry.EndedAt = &end

	if err := entry.Save(context.Background()); err != nil {
		return nil, err
	}
	return entry, nil
}

func (p *Project) CreateEntry(content string, billable bool) (*Entry, error) {
	entry := &Entry{
		db:        p.db,
		Project:   p,
		ProjectID: p.ID,
		Content:   strings.TrimSpace(content),
		Type:      EntryTypeWork,
		Billable:  billable,
		Tags:      extractTags(content),
	}
	start := time.Now().UTC()
	entry.StartedAt = &start
	entry.DurationMs = 0

	if err := entry.Save(context.Background()); err != nil {
		return nil, err
	}
	return entry, nil
}

func (p *Project) OnClock() (bool, *Timer) {
	timer, err := p.currentTimer()
	if err != nil || timer == nil {
		return false, nil
	}
	return true, timer
}

func (p *Project) currentTimer() (*Timer, error) {
	ctx := context.Background()
	record, err := p.db.queries.GetTimer(ctx, p.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetch timer: %w", err)
	}
	return newTimerFromModel(p.db, p, record), nil
}

func (p *Project) Delete() error {
	if p == nil || p.db == nil {
		return errors.New("project not initialized")
	}
	if err := p.db.queries.DeleteProject(context.Background(), p.ID); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}

func (p *Project) Rename(newName string) error {
	if p == nil || p.db == nil {
		return errors.New("project not initialized")
	}
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return errors.New("project name cannot be empty")
	}
	if strings.EqualFold(newName, p.Name) {
		return nil
	}
	record, err := p.db.queries.UpdateProject(context.Background(), sqlc.UpdateProjectParams{
		ID:       p.ID,
		Name:     newName,
		Company:  optionalString(p.Company),
		IsHidden: boolToInt(p.IsHidden),
	})
	if err != nil {
		return fmt.Errorf("rename project: %w", err)
	}
	p.Name = record.Name
	if record.Company.Valid {
		value := record.Company.String
		p.Company = &value
	} else {
		p.Company = nil
	}
	p.IsHidden = record.IsHidden == 1
	p.UpdatedAt = time.Unix(record.UpdatedAt, 0).UTC()
	return nil
}

func optionalString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func (d *Database) Projects() []*Project {
	if d == nil {
		return nil
	}
	ctx := context.Background()
	rows, err := d.queries.ListProjects(ctx)
	if err != nil {
		fmt.Printf("Failed to list projects: %v\n", err)
		return nil
	}
	projects := make([]*Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, newProjectFromModel(d, row))
	}
	sort.Slice(projects, func(i, j int) bool {
		return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name)
	})
	return projects
}

func (d *Database) CreateProject(name string) (*Project, error) {
	if d == nil {
		return nil, errors.New("database not initialized")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("project name cannot be empty")
	}

	record, err := d.queries.CreateProject(context.Background(), sqlc.CreateProjectParams{
		Name:     name,
		Company:  sql.NullString{},
		IsHidden: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return newProjectFromModel(d, record), nil
}
