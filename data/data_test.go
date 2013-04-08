package data_test

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/nexneo/samay/data"
	"os"
	"testing"
	"time"
)

func TestProjectCreation(t *testing.T) {
	t.Parallel()
	project := data.CreateProject("Test 1")
	project.Company = proto.String("Insight Methods, Inc.")
	if err := data.Save(project); err != nil {
		t.Error(err.Error())
	}
}

func TestTimerCreation(t *testing.T) {
	t.Parallel()
	project := data.CreateProject("Test 2")
	timer := data.CreateTimer(project)
	<-time.After(time.Microsecond)
	if *timer.Started == 0 || *timer.Started < 0 {
		t.Error("timer Started should be greater then 0")
	}
	if err := data.Save(timer); err != nil {
		t.Error(err.Error())
	}
	os.RemoveAll(data.DB.ProjectDirPath(project))
}

func TestTimerStop(t *testing.T) {
	t.Parallel()
	project := data.CreateProject("Test 3")
	timer := data.CreateTimer(project)
	if *timer.Started == 0 || *timer.Started < 0 {
		t.Error("timer Started should be greater then 0")
	}

	// timer to test actual duration stored in entry
	counter := time.After(time.Second)

	// create entry to pass into timer
	e := project.CreateEntry("Working on some stuff", true)

	// wait for second and then stop
	<-counter
	if err := timer.Stop(e); err != nil {
		t.Error(err.Error())
	}
	if time.Duration(e.GetDuration()) < time.Second {
		t.Errorf(
			"entry duration should: %v > %v",
			time.Duration(e.GetDuration()), time.Second,
		)
	}
	started, _ := e.StartedTime()
	ended, _ := e.EndedTime()
	if started.Sub(*ended) == time.Duration(e.GetDuration()) {
		t.Errorf(
			"entry duration should be correct. %v",
			e.GetDuration(),
		)
	}
	os.RemoveAll(data.DB.ProjectDirPath(project))
}

func TestTimerLoad(t *testing.T) {
	t.Parallel()
	project := data.CreateProject("Test 4")
	timer := data.CreateTimer(project)

	<-time.After(time.Second)
	if err := data.Save(timer); err != nil {
		t.Error(err.Error())
	}

	<-time.After(10 * time.Millisecond)
	timer2 := data.GetTimer(project)
	if timer2.GetStarted() != timer.GetStarted() {
		t.Fail()
	}
	os.RemoveAll(data.DB.ProjectDirPath(project))
}
