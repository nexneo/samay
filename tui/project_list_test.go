package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/nexneo/samay/data"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "samay-tui-test-*")
	if err != nil {
		panic(fmt.Sprintf("create temp dir: %v", err))
	}
	path := filepath.Join(dir, "test.db")
	if err := data.OpenDatabase(path); err != nil {
		panic(fmt.Sprintf("open database: %v", err))
	}

	code := m.Run()

	if data.DB != nil {
		_ = data.DB.Close()
	}
	_ = os.RemoveAll(dir)

	os.Exit(code)
}

func TestHandleNumericProjectSelection_AllowsImmediateReuse(t *testing.T) {
	a := newTestApp(t, []string{"Alpha", "Bravo", "Charlie"})

	if !a.handleNumericProjectSelection("3") {
		t.Fatalf("expected to handle key 3")
	}
	if idx := a.projects.Index(); idx != 2 {
		t.Fatalf("expected selection index 2 after pressing 3, got %d", idx)
	}

	if !a.handleNumericProjectSelection("2") {
		t.Fatalf("expected to handle key 2")
	}
	if idx := a.projects.Index(); idx != 1 {
		t.Fatalf("expected selection index 1 after pressing 2, got %d", idx)
	}
}

func TestHandleNumericProjectSelection_MultiDigitStillWorks(t *testing.T) {
	names := make([]string, 15)
	for i := range names {
		names[i] = fmt.Sprintf("Project %02d", i+1)
	}
	a := newTestApp(t, names)

	if !a.handleNumericProjectSelection("1") {
		t.Fatalf("expected to handle key 1")
	}
	if idx := a.projects.Index(); idx != 0 {
		t.Fatalf("expected selection index 0 after pressing 1, got %d", idx)
	}

	if !a.handleNumericProjectSelection("2") {
		t.Fatalf("expected to handle key 2")
	}
	if idx := a.projects.Index(); idx != 11 {
		t.Fatalf("expected selection index 11 after pressing 12, got %d", idx)
	}

	if !a.handleNumericProjectSelection("3") {
		t.Fatalf("expected to handle key 3")
	}
	if idx := a.projects.Index(); idx != 2 {
		t.Fatalf("expected selection index 2 after pressing 3, got %d", idx)
	}
}

func newTestApp(t *testing.T, names []string) *app {
	t.Helper()

	clearProjects(t)

	for _, name := range names {
		if _, err := data.DB.CreateProject(name); err != nil {
			t.Fatalf("create project %q: %v", name, err)
		}
	}

	items := make([]list.Item, len(names))
	for i, name := range names {
		items[i] = item(name)
	}

	height := len(items)*2 + 5
	if height < 10 {
		height = 10
	}

	l := list.New(items, itemDelegate{}, 40, height)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	a := &app{
		projects: l,
		state:    stateProjectList,
	}
	a.updateProjectSelectionFromList()

	return a
}

func clearProjects(t *testing.T) {
	t.Helper()
	projects := data.DB.Projects()
	for _, project := range projects {
		if err := project.Delete(); err != nil {
			t.Fatalf("delete project %q: %v", project.Name, err)
		}
	}
}
