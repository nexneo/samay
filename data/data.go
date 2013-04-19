package data

import (
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"fmt"
	"github.com/nexneo/samay/util"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	_ = fmt.Errorf
	_ = errors.New
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

func (p *Project) Location() string {
	return DB.ProjectDirPath(p) + "/project.db"
}

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

func (p *Project) Entries() (entries []*Entry) {
	files, err := util.ReadDir(DB.ProjectDirPath(p) + "/entries")
	if err != nil {
		return
	}

	for _, file := range files {
		entry := new(Entry)
		entry.Project = p
		entry.Id = proto.String(file.Name())
		if err = Load(entry); err == nil {
			entries = append(entries, entry)
		} else {
			fmt.Println(err)
		}
	}
	return
}

// Timer

func CreateTimer(project *Project) *Timer {
	timer := new(Timer)
	timer.Project = project
	timer.Start()
	return timer
}

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

func (t *Timer) Location() string {
	return DB.ProjectDirPath(t.GetProject()) + "/timer.db"
}

func (t *Timer) Start() {
	t.Started = proto.Int64(time.Now().Unix())
}

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

func (t *Timer) StartedTime() (*time.Time, error) {
	return util.TimestampToTime(t.Started)
}

func (t *Timer) Duration() time.Duration {
	v, _ := t.StartedTime()
	return time.Now().Sub(*v)
}

// Entry

func (project *Project) CreateEntry(content string, billable bool) *Entry {
	content = strings.Trim(content, " \n\t\r")
	tags := make([]string, 0, 20)

	tagsFinder := regexp.MustCompile("\\B#(\\w\\w+)")
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

func (e *Entry) prepareForSave() error {
	if err := os.MkdirAll(DB.EntryDirPath(e), 0755); !os.IsExist(err) {
		return err
	}
	return nil
}

func (p *Entry) cleanupAfterRemove() error {
	return nil
}

func (e *Entry) Location() string {
	return DB.EntryDirPath(e) + "/" + e.GetId()
}

func (e *Entry) StartedTime() (*time.Time, error) {
	return util.TimestampToTime(e.Started)
}

func (e *Entry) EndedTime() (*time.Time, error) {
	return util.TimestampToTime(e.Ended)
}

func (e *Entry) Minutes() float64 {
	return time.Duration(*e.Duration).Minutes()
}

func (e *Entry) HoursMins() hoursMins {
	d := time.Duration(*e.Duration)
	return hoursMinsFromDuration(d)
}
