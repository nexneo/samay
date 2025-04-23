package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nexneo/samay/data"
	"github.com/samber/lo"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type app struct {
	project  *data.Project
	projects list.Model
	choices  [][2]string
}

func CreateApp() *app {
	items := lo.Map(data.DB.Projects(), func(p *data.Project, _ int) list.Item {
		name := *p.Name
		return item(name)
	})
	const defaultWidth = 20
	const listHeight = 14

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Please choose a project"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &app{
		project:  nil,
		projects: l,
		choices: [][2]string{
			{"s", "Start timer"},
			{"p", "End timer"},
			{"e", "Enter manually"},
			{"", "Show logs"},
			{"", "Edit project"},
			{"", "Delete project"},
		},
	}
}

func (a app) Init() tea.Cmd {

	return nil
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.projects.SetWidth(msg.Width)
		return a, nil

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch keypress := msg.String(); keypress {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return a, tea.Quit

		case "s":
			a.project.StartTimer()
			return a, nil
		case "p":
			a.project.StopTimer("", true)
			return a, nil
		case "enter":
			i, ok := a.projects.SelectedItem().(item)
			if ok {
				a.project = data.CreateProject(string(i))
			}
			return a, tea.Quit

		}
	}

	var cmd tea.Cmd
	a.projects, cmd = a.projects.Update(msg)
	return a, cmd
}

func (a app) View() string {
	if a.project == nil {
		return a.projects.View()
	}
	// Display the choices in a simple list format
	var output string
	output += "Please choose an option: " + *a.project.Name + "\n\n"
	for _, choice := range a.choices {
		output += fmt.Sprintf("[%s] %s\n", choice[0], choice[1])
	}
	output += "\nPress q to quit.\n"
	return quitTextStyle.Render(output)
}
