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
)

type app struct {
	project  *data.Project
	projects list.Model
	choices  [][2]string
}

func CreateApp() *app {
	var project *data.Project
	projects := data.DB.Projects()
	items := lo.Map(projects, func(p *data.Project, _ int) list.Item {
		name := *p.Name
		return item(name)
	})
	const defaultWidth = 20
	listHeight := len(items)*2 + 5

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Please choose a project"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.TitleBar.Padding(3)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	for _, p := range projects {
		timer, _ := p.OnClock()
		if timer {
			project = p
			break
		}
	}

	return &app{
		project:  project,
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
	case tea.KeyMsg:
		m, c := a.handleKeypress(msg)
		if m != nil {
			return m, c
		}
	}

	var cmd tea.Cmd
	a.projects, cmd = a.projects.Update(msg)
	return a, cmd
}

func (a app) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "esc":
		a.project = nil
		return a, nil
	case "ctrl+c", "q":
		return a, tea.Quit
	case "s":
		a.project.StartTimer()
		return a, nil
	case "p":
		a.project.StopTimer("", true)
		return a, nil
	case "enter":
		if i, ok := a.projects.SelectedItem().(item); ok {
			a.project = data.CreateProject(string(i))
		}
		return a, nil
	default:
		return nil, nil
	}
}

func (a app) View() string {
	if a.project == nil {
		return a.projects.View()
	}

	var lines []string

	headerText := "Choose an option for project: " + *a.project.Name + "\n" // Added newline here
	onclock, _ := a.project.OnClock()
	if onclock {
		headerText = "Timer is running for project: " + *a.project.Name + "\n" // Added newline here
	}

	lines = append(lines, titleStyle.MarginTop(1).Render(headerText))

	// Render each choice as a separate line
	for _, choice := range a.choices {
		if onclock && choice[0] == "s" {
			continue
		}
		choiceText := fmt.Sprintf("[%s] %s", choice[0], choice[1])
		lines = append(lines, itemStyle.Render(choiceText))
	}

	// Render the help text separately
	helpText := "Press q to quit." // Leading newline for spacing
	lines = append(lines, helpStyle.Render(helpText))

	// Join all the rendered lines vertically
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
