package data

import (
	"testing"
	"time"

	"github.com/nexneo/samay/data/sqlc"
)

func TestNewTimerFromModel(t *testing.T) {
	start := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	created := start.Add(30 * time.Second)
	updated := created.Add(15 * time.Second)

	model := sqlc.Timer{
		ProjectID: 42,
		StartedAt: start.Unix(),
		CreatedAt: created.Unix(),
		UpdatedAt: updated.Unix(),
	}

	timer := newTimerFromModel(nil, nil, model)
	if timer.ProjectID != 42 {
		t.Fatalf("expected project ID 42, got %d", timer.ProjectID)
	}
	if !timer.StartedAt.Equal(start) {
		t.Fatalf("expected started time %v, got %v", start, timer.StartedAt)
	}
	if !timer.CreatedAt.Equal(created) {
		t.Fatalf("expected created time %v, got %v", created, timer.CreatedAt)
	}
	if !timer.UpdatedAt.Equal(updated) {
		t.Fatalf("expected updated time %v, got %v", updated, timer.UpdatedAt)
	}
}

func TestTimerAccessorsAndDuration(t *testing.T) {
	var timer *Timer
	if timer.GetStarted() != 0 {
		t.Fatalf("expected zero started timestamp for nil timer")
	}
	if !timer.StartedTime().IsZero() {
		t.Fatalf("expected zero time for nil timer")
	}
	if timer.Duration() != 0 {
		t.Fatalf("expected zero duration for nil timer")
	}

	timer = &Timer{}
	if timer.Duration() != 0 {
		t.Fatalf("expected zero duration for timer without start")
	}

	start := time.Now().Add(-1500 * time.Millisecond)
	timer = &Timer{StartedAt: start}
	duration := timer.Duration()
	if duration < time.Second || duration > 3*time.Second {
		t.Fatalf("expected duration between 1s and 3s, got %v", duration)
	}
	if timer.GetStarted() != start.Unix() {
		t.Fatalf("expected started unix %d, got %d", start.Unix(), timer.GetStarted())
	}
	if !timer.StartedTime().Equal(start) {
		t.Fatalf("expected started time %v, got %v", start, timer.StartedTime())
	}
}
