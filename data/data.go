package data

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nexneo/samay/util"
	"google.golang.org/protobuf/proto"
)

// Types

type Persistable interface {
	locator
	proto.Message
	prepareForSave() error
}

type Destroyable interface {
	locator
	cleanupAfterRemove() error
}

type locator interface {
	Location() string
}

// Project

// CreateProject returns a new in-memory project proto tagged with the provided name.
// The caller is responsible for persisting it via Save to materialize it on disk.
func CreateProject(name string) *Project {
	project := new(Project)
	project.Name = proto.String(name)
	return project
}

func (p *Project) prepareForSave() error {
	return DB.MkProjectDir(p)
}

func (p *Project) cleanupAfterRemove() error {
	return os.RemoveAll(DB.ProjectDirPath(p))
}

// Location resolves the project database path underneath the configured Dropbox root.
func (p *Project) Location() string {
	return DB.ProjectDirPath(p) + "/project.db"
}

// GetShaFromName returns a stable directory identifier for the project backed by
// the configured name or falls back to an existing SHA/empty-name hash when needed.
func (p *Project) GetShaFromName() string {
	if p.GetName() != "" {
		return util.SHA1(p.GetName())
	}
	// Then try if Sha is set, from directory name.
	if sha := p.GetSha(); sha != "" {
		return sha
	}
	// mostly like Project name is not set, use default.
	return util.SHA1("")
}

type byEndTime []*Entry

func (f byEndTime) Len() int      { return len(f) }
func (f byEndTime) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func (f byEndTime) Less(i, j int) bool {
	return *f[i].Ended > *f[j].Ended
}

// Entries returns the project's persisted entries ordered from newest to oldest.
// Any storage errors are logged so the caller can keep rendering partial results.
func (p *Project) Entries() (entries []*Entry) {
	entries, err := DB.repos.entries.ForProject(p)
	if err != nil {
		fmt.Println("Failed to list entries:", err)
	}
	return entries
}

// Timer

// CreateTimer prepares a new timer for the project and eagerly marks it as started.
func CreateTimer(project *Project) *Timer {
	timer := new(Timer)
	timer.Project = project
	timer.Start()
	return timer
}

// GetTimer loads the active timer for the project, creating a new stopped one if none exists.
func GetTimer(project *Project) *Timer {
	timer := new(Timer)
	timer.Project = project
	err := Load(timer)
	if err != nil {
		timer = CreateTimer(project)
		timer.Started = proto.Int64(0)
	}
	return timer
}

func (t *Timer) prepareForSave() error {
	return DB.MkProjectDir(t.GetProject())
}

func (p *Timer) cleanupAfterRemove() error {
	return nil
}

// Location resolves where the timer proto is persisted for its project.
func (t *Timer) Location() string {
	return DB.ProjectDirPath(t.GetProject()) + "/timer.db"
}

// Start records the current wall clock as the timer's start time.
func (t *Timer) Start() {
	t.Started = proto.Int64(time.Now().Unix())
}

// Stop finalizes the timer by materializing the supplied entry and clearing the timer state.
// It persists the entry, tears down the timer, and returns the first error encountered.
func (t *Timer) Stop(e *Entry) error {
	stopped := time.Now()
	s, err := t.StartedTime()
	if err != nil {
		return err
	}
	e.Project = t.GetProject()
	e.Started = t.Started
	e.Ended = proto.Int64(stopped.Unix())
	e.Duration = proto.Int64(stopped.Sub(*s).Nanoseconds())
	if err = Save(e); err != nil {
		return err
	}
	if err = Destroy(t); err != nil {
		return err
	}
	return nil
}

// StartedTime converts the stored UNIX timestamp into a Go time for consumption by callers.
func (t *Timer) StartedTime() (*time.Time, error) {
	return util.TimestampToTime(t.Started)
}

// Duration reports how long the timer has been running based on the captured start time.
func (t *Timer) Duration() time.Duration {
	v, _ := t.StartedTime()
	return time.Since(*v)
}

// Entry

// CreateEntry constructs a new entry for the project, extracting inline #tags while trimming content.
// The returned entry is in-memory only until persisted via Save.
func (project *Project) CreateEntry(content string, billable bool) *Entry {
	content = strings.Trim(content, " \n\t\r")
	tags := make([]string, 0, 20)

	tagsFinder := regexp.MustCompile(`\B#(\w\w+)`)
	for _, v := range tagsFinder.FindAllStringSubmatch(content, 20) {
		if len(v) > 1 {
			tags = append(tags, v[1])
		}
	}
	e := Entry{
		Project:  project,
		Tags:     tags,
		Content:  proto.String(content),
		Billable: proto.Bool(billable),
	}
	id, _ := util.UUID()
	e.Id = proto.String(id)
	return &e
}

// CreateEntryWithDuration backfills a synthetic entry with the provided duration instead of the live clock.
func (project *Project) CreateEntryWithDuration(
	content string,
	duration time.Duration,
	billable bool) *Entry {

	e := project.CreateEntry(content, billable)
	endTime := time.Now()
	startTime := endTime.Add(-duration)

	e.Started = proto.Int64(startTime.Unix())
	e.Ended = proto.Int64(endTime.Unix())
	e.Duration = proto.Int64(duration.Nanoseconds())

	return e
}

// StopTimer ends the running timer, saves the captured work entry, and echoes the duration to stdout.
func (project *Project) StopTimer(c string, bill bool) (err error) {
	if yes, timer := project.OnClock(); yes {
		entry := project.CreateEntry(c, bill)

		if err = timer.Stop(entry); err == nil {
			fmt.Printf("%.2f mins\n", entry.Minutes())
		}
	}
	return
}

// StartTimer persists the project if needed and records a fresh timer instance on disk.
func (project *Project) StartTimer() (err error) {
	Save(project)
	timer := CreateTimer(project)
	err = Save(timer)
	return
}

func (e *Entry) prepareForSave() error {
	if err := os.MkdirAll(DB.EntryDirPath(e), 0755); !os.IsExist(err) {
		return err
	}
	return nil
}

func (p *Entry) cleanupAfterRemove() error {
	return nil
}

// Location returns the filesystem path backing the entry payload inside its project.
func (e *Entry) Location() string {
	return DB.EntryDirPath(e) + "/" + e.GetId()
}

// StartedTime exposes the entry's start timestamp as a Go time, returning parse errors verbatim.
func (e *Entry) StartedTime() (*time.Time, error) {
	return util.TimestampToTime(e.Started)
}

// EndedTime exposes the entry's end timestamp as a Go time, returning parse errors verbatim.
func (e *Entry) EndedTime() (*time.Time, error) {
	return util.TimestampToTime(e.Ended)
}

// Minutes reports the entry duration expressed in fractional minutes for display helpers.
func (e *Entry) Minutes() float64 {
	return time.Duration(*e.Duration).Minutes()
}

// HoursMins returns the entry duration split into hour/minute components.
func (e *Entry) HoursMins() hoursMins {
	d := time.Duration(*e.Duration)
	return hoursMinsFromDuration(d)
}
