package data

import (
	"time"

	"github.com/nexneo/samay/data/sqlc"
)

type Timer struct {
	db        *Database
	Project   *Project
	ProjectID int64
	StartedAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func newTimerFromModel(db *Database, project *Project, model sqlc.Timer) *Timer {
	return &Timer{
		db:        db,
		Project:   project,
		ProjectID: model.ProjectID,
		StartedAt: time.Unix(model.StartedAt, 0).UTC(),
		CreatedAt: time.Unix(model.CreatedAt, 0).UTC(),
		UpdatedAt: time.Unix(model.UpdatedAt, 0).UTC(),
	}
}

func (t *Timer) GetStarted() int64 {
	if t == nil {
		return 0
	}
	return t.StartedAt.Unix()
}

func (t *Timer) StartedTime() time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.StartedAt
}

func (t *Timer) Duration() time.Duration {
	if t == nil {
		return 0
	}
	start := t.StartedAt
	if start.IsZero() {
		return 0
	}
	return time.Since(start)
}
