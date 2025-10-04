package data_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nexneo/samay/data"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getwd: %v\n", err)
		os.Exit(1)
	}

	testDataPath := filepath.Join(wd, "..", "testing_data")
	fmt.Println("Using test data path:", testDataPath)
	if err := resetTestDataDir(testDataPath); err != nil {
		fmt.Fprintf(os.Stderr, "prepare test data dir: %v\n", err)
		os.Exit(1)
	}

	if err := data.SetBasePath(testDataPath); err != nil {
		fmt.Fprintf(os.Stderr, "set base path: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := resetTestDataDir(testDataPath); err != nil {
		fmt.Fprintf(os.Stderr, "cleanup test data dir: %v\n", err)
		os.Exit(1)
	}

	os.Exit(code)
}

func resetTestDataDir(path string) error {
	fmt.Println("Resetting test data directory:", path)
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	if err := os.MkdirAll(path, 0775); err != nil {
		return err
	}
	gitkeep := filepath.Join(path, ".gitkeep")
	if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
		if err := os.WriteFile(gitkeep, []byte{}, 0666); err != nil {
			return err
		}
	}
	return nil
}

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
