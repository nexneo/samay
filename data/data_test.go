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
	if err := os.RemoveAll(data.DB.ProjectDirPath(project)); err != nil {
		t.Fatalf("cleanup project dir: %v", err)
	}
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
	if err := os.RemoveAll(data.DB.ProjectDirPath(project)); err != nil {
		t.Fatalf("cleanup project dir: %v", err)
	}
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
	if err := os.RemoveAll(data.DB.ProjectDirPath(project)); err != nil {
		t.Fatalf("cleanup project dir: %v", err)
	}
}

func TestDestroyRemovesProjectDirectory(t *testing.T) {
	project := data.CreateProject("Test Cleanup")
	if err := data.Save(project); err != nil {
		t.Fatal(err)
	}
	dir := data.DB.ProjectDirPath(project)
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected project directory, got %v", err)
	}
	if err := data.Destroy(project); err != nil {
		t.Fatalf("destroy project: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("expected directory removal, got: %v", err)
	}
}

func TestProjectEntriesReturnsSortedEntries(t *testing.T) {
	project := data.CreateProject("Test Entries")
	now := time.Now().Add(-5 * time.Minute)
	for i := 0; i < 3; i++ {
		entry := project.CreateEntry(fmt.Sprintf("log %d", i), true)
		end := now.Add(time.Duration(i) * time.Minute)
		start := end.Add(-time.Minute)
		entry.Started = proto.Int64(start.Unix())
		entry.Ended = proto.Int64(end.Unix())
		entry.Duration = proto.Int64(time.Minute.Nanoseconds())
		if err := data.Save(entry); err != nil {
			t.Fatalf("save entry %d: %v", i, err)
		}
	}
	entriesDir := filepath.Join(data.DB.ProjectDirPath(project), "entries")
	if err := os.WriteFile(filepath.Join(entriesDir, ".DS_Store"), []byte("ignore"), 0644); err != nil {
		t.Fatalf("create sentinel file: %v", err)
	}
	entries := project.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for i := 1; i < len(entries); i++ {
		prevEnded, err := entries[i-1].EndedTime()
		if err != nil {
			t.Fatalf("previous entry ended time: %v", err)
		}
		currEnded, err := entries[i].EndedTime()
		if err != nil {
			t.Fatalf("current entry ended time: %v", err)
		}
		if prevEnded.Before(*currEnded) {
			t.Fatalf("entries not sorted by descending end time: %v before %v", prevEnded, currEnded)
		}
	}
	if err := os.RemoveAll(data.DB.ProjectDirPath(project)); err != nil {
		t.Fatalf("cleanup project dir: %v", err)
	}
}

func TestTimerDuration(t *testing.T) {
	project := data.CreateProject("Test Duration")
	timer := data.CreateTimer(project)
	timer.Started = proto.Int64(time.Now().Add(-time.Minute).Unix())
	if elapsed := timer.Duration(); elapsed < time.Minute {
		t.Fatalf("expected at least 1 minute elapsed, got %v", elapsed)
	}
	if err := os.RemoveAll(data.DB.ProjectDirPath(project)); err != nil {
		t.Fatalf("cleanup project dir: %v", err)
	}
}

func TestCreateEntryWithDuration(t *testing.T) {
	project := data.CreateProject("Test Entry Duration")
	duration := 15 * time.Minute
	entry := project.CreateEntryWithDuration(" working \n", duration, true)
	if entry.GetContent() != "working" {
		t.Fatalf("expected trimmed content, got %q", entry.GetContent())
	}
	if got := time.Duration(entry.GetDuration()); got != duration {
		t.Fatalf("expected duration %v, got %v", duration, got)
	}
	started, err := entry.StartedTime()
	if err != nil {
		t.Fatalf("started time: %v", err)
	}
	ended, err := entry.EndedTime()
	if err != nil {
		t.Fatalf("ended time: %v", err)
	}
	if ended.Sub(*started) != duration {
		t.Fatalf("expected time difference %v, got %v", duration, ended.Sub(*started))
	}
}

func TestProjectStartAndStopTimer(t *testing.T) {
	project := data.CreateProject("Test Start Stop")
	if err := project.StartTimer(); err != nil {
		t.Fatalf("start timer: %v", err)
	}
	if onClock, timer := project.OnClock(); !onClock || timer.GetStarted() == 0 {
		t.Fatalf("expected timer on clock, got onClock=%v started=%d", onClock, timer.GetStarted())
	}
	time.Sleep(25 * time.Millisecond)
	if err := project.StopTimer("doing work", true); err != nil {
		t.Fatalf("stop timer: %v", err)
	}
	if onClock, timer := project.OnClock(); onClock || timer.GetStarted() != 0 {
		t.Fatalf("expected timer cleared, got onClock=%v started=%d", onClock, timer.GetStarted())
	}
	entries := project.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].GetContent() != "doing work" {
		t.Fatalf("unexpected entry content: %q", entries[0].GetContent())
	}
	if entries[0].GetBillable() != true {
		t.Fatalf("expected billable entry")
	}
	if _, err := os.Stat(data.DB.ProjectDirPath(project) + "/timer.db"); !os.IsNotExist(err) {
		t.Fatalf("expected timer file removed, got: %v", err)
	}
	if err := os.RemoveAll(data.DB.ProjectDirPath(project)); err != nil {
		t.Fatalf("cleanup project dir: %v", err)
	}
}

func TestEntryDurationHelpers(t *testing.T) {
	project := data.CreateProject("Test Helpers")
	entry := project.CreateEntry("helper", false)
	entry.Duration = proto.Int64((90 * time.Minute).Nanoseconds())
	if minutes := entry.Minutes(); minutes != 90 {
		t.Fatalf("expected 90 minutes, got %v", minutes)
	}
	if hm := entry.HoursMins(); hm.String() != "1:30" {
		t.Fatalf("expected formatted string '1:30', got %q", hm.String())
	}
}

func TestPersistedAndUpdate(t *testing.T) {
	project := data.CreateProject("Test Persisted")
	if data.Persisted(project) {
		t.Fatalf("expected project to be absent before save")
	}
	if err := data.Save(project); err != nil {
		t.Fatalf("save project: %v", err)
	}
	if !data.Persisted(project) {
		t.Fatalf("expected project to be persisted")
	}
	project.Company = proto.String("Acme Co")
	if err := data.Update(project); err != nil {
		t.Fatalf("update project: %v", err)
	}
	reloaded := &data.Project{}
	reloaded.Sha = proto.String(project.GetShaFromName())
	if err := data.Load(reloaded); err != nil {
		t.Fatalf("reload project: %v", err)
	}
	if reloaded.GetCompany() != "Acme Co" {
		t.Fatalf("expected updated company, got %q", reloaded.GetCompany())
	}
	if err := os.RemoveAll(data.DB.ProjectDirPath(project)); err != nil {
		t.Fatalf("cleanup project dir: %v", err)
	}
}

func TestDropboxString(t *testing.T) {
	if data.DB.String() == "" {
		t.Fatalf("expected dropbox base path to be set")
	}
}
