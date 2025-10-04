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

	ended, _ := entry.EndedTime()
	endedStr := "--"
	if ended != nil && !ended.IsZero() {
		endedStr = ended.Format("2006-01-02 15:04")
	}

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

	line := fmt.Sprintf("%2d %s %s %s", index+1, hm, endedStr, desc)

	if index == m.Index() {
		highlight := selectedItemStyle.Copy().PaddingLeft(4)
		fmt.Fprint(w, highlight.Render(line))
		return
	}

	fmt.Fprint(w, itemStyle.Render(line))
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
