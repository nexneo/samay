package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nexneo/samay/data"
)

type entryItem struct {
	entry *data.Entry
}

func (i entryItem) FilterValue() string {
	if i.entry == nil {
		return ""
	}
	return strings.ToLower(i.entry.GetContent())
}

type entryItemDelegate struct{}

func (d entryItemDelegate) Height() int                             { return 1 }
func (d entryItemDelegate) Spacing() int                            { return 0 }
func (d entryItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d entryItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it, ok := listItem.(entryItem)
	if !ok || it.entry == nil {
		return
	}

	entry := it.entry
	hm := data.HmFromD(time.Duration(entry.GetDuration()))

	desc := strings.TrimSpace(strings.ReplaceAll(entry.GetContent(), "\n", " "))
	if len(desc) == 0 {
		desc = "(no description)"
	}

	maxDesc := 60
	if maxDesc > m.Width()-24 {
		maxDesc = m.Width() - 24
	}
	if maxDesc < 20 {
		maxDesc = 20
	}

	if len(desc) > maxDesc {
		desc = desc[:maxDesc-3] + "..."
	}

	line := fmt.Sprintf("%2d %s %s", index+1, hm, desc)

	if index == m.Index() {
		highlight := selectedItemStyle.PaddingLeft(4)
		if _, err := fmt.Fprint(w, highlight.Render(line)); err != nil {
			return
		}
		return
	}

	_, _ = fmt.Fprint(w, itemStyle.Render(line))
}

func buildEntryList(project *data.Project, width, height int) list.Model {
	entries := project.Entries()
	items := make([]list.Item, 0, len(entries))
	for _, e := range entries {
		items = append(items, entryItem{entry: e})
	}

	if height <= 0 {
		height = 10
	}
	l := list.New(items, entryItemDelegate{}, width, height)
	l.Title = ""
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(true)
	l.SetShowStatusBar(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	return l
}

func entryFromListItem(i list.Item) *data.Entry {
	if it, ok := i.(entryItem); ok {
		return it.entry
	}
	return nil
}

func (a *app) handleKeypressEntryList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "esc":
		a.state = stateProjectMenu
		a.errorMessage = ""
		return a, nil
	case "m":
		entry := entryFromListItem(a.entries.SelectedItem())
		if entry == nil {
			return a, nil
		}
		a.selectedEntry = entry
		a.prepareMoveProjectList()
		if len(a.moveProjects.Items()) == 0 {
			a.errorMessage = "No other projects available to move this entry."
			return a, nil
		}
		a.previousState = stateEntryList
		a.state = stateMoveEntryTarget
		return a, nil
	case "d":
		entry := entryFromListItem(a.entries.SelectedItem())
		if entry != nil {
			a.selectedEntry = entry
			a.confirmEntry = entry
			a.confirmAction = confirmDeleteEntry
			a.confirmMessage = fmt.Sprintf("Delete entry '%s'?", truncateString(entry.GetContent(), 40))
			a.previousState = stateEntryList
			a.state = stateConfirm
		}
		return a, nil
	}

	var cmd tea.Cmd
	a.entries, cmd = a.entries.Update(msg)
	a.selectedEntry = entryFromListItem(a.entries.SelectedItem())
	return a, cmd
}
