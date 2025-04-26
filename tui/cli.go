package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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
	inputPromptStyle  = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("240")) // Style for input prompt
)

// Define different states for the application
type state int

const (
	stateProjectList   state = iota // Showing the list of projects
	stateProjectMenu                // Showing the menu for a selected project
	stateStoppingTimer              // Asking for a message before stopping timer
)

type app struct {
	project  *data.Project
	projects list.Model
	choices  [][2]string
	state    state
	message  textinput.Model // input for stop message
	width    int             // store window width
}

func CreateApp() *app {
	var currentProject *data.Project
	projects := data.DB.Projects()
	items := lo.Map(projects, func(p *data.Project, _ int) list.Item {
		name := *p.Name
		return item(name)
	})
	const defaultWidth = 20
	listHeight := len(items)*2 + 5 // Adjust height based on items

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Please choose a project"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.TitleBar.Padding(3)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// Check if any project has a running timer
	for _, p := range projects {
		timer, _ := p.OnClock()
		if timer {
			currentProject = p
			break
		}
	}

	// text input model
	ti := textinput.New()
	ti.Placeholder = "Enter stop message (optional, press Enter to submit, Esc to cancel)"
	ti.CharLimit = 156
	ti.Width = 50 // Adjust width as needed

	initialState := stateProjectList
	if currentProject != nil {
		// show project menu if a timer is running, otherwise show project list
		initialState = stateProjectMenu
	}

	return &app{
		project:  currentProject,
		projects: l,
		state:    initialState,
		choices: [][2]string{
			{"s", "Start timer"},
			{"p", "End timer"},
			{"e", "Enter manually"},
			{"", "Show logs"},
			{"", "Edit project"},
			{"", "Delete project"},
		},
		message: ti,
	}
}

func (a app) Init() tea.Cmd {
	return nil // no initial command for now
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd // Slice to hold commands for project menu

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width // Store width
		a.projects.SetWidth(msg.Width)
		// a.messageInput.Width = msg.Width - 10 // adjustment
		return a, nil

	case tea.KeyMsg:
		// Handle key presses based on the current state
		switch a.state {
		case stateProjectList:
			m, c := a.handleKeypressProjectList(msg)
			return m, c
		case stateProjectMenu:
			m, c := a.handleKeypressProjectMenu(msg)
			return m, c
		case stateStoppingTimer:
			m, c := a.handleKeypressStoppingTimer(msg)
			return m, c
		}
	}

	// If not handled by state-specific keypress logic, update the current model
	switch a.state {
	case stateProjectList:
		a.projects, cmd = a.projects.Update(msg)
		cmds = append(cmds, cmd)
	case stateStoppingTimer:
		a.message, cmd = a.message.Update(msg)
		cmds = append(cmds, cmd)
		// No default update for project menu state needed here
	}

	return a, tea.Batch(cmds...) // Batch commands
}

// when the project list is active
func (a *app) handleKeypressProjectList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "enter":
		if i, ok := a.projects.SelectedItem().(item); ok {
			selectedProj := data.CreateProject(string(i))
			if selectedProj != nil {
				a.project = selectedProj
				a.state = stateProjectMenu
			}
		}
		return a, nil
	}

	// Default list navigation
	var cmd tea.Cmd
	a.projects, cmd = a.projects.Update(msg)
	return a, cmd
}

// when the project menu is active
func (a *app) handleKeypressProjectMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.project == nil { // Should not happen in this state, but safety check
		a.state = stateProjectList
		return a, nil
	}

	onclock, _ := a.project.OnClock()

	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "esc": // Go back to project list
		a.project = nil
		a.state = stateProjectList
		return a, nil
	case "s":
		if !onclock {
			err := a.project.StartTimer()
			if err != nil {
				// Handle error (e.g., display a message)
				fmt.Println("Error starting timer:", err) // Simple console log for now
			}
			// No state change needed, view will update
		}
		return a, nil
	case "p":
		if onclock {
			a.state = stateStoppingTimer // Switch to stopping state
			a.message.Focus()            // Focus the text input
			a.message.SetValue("")       // Clear previous message
			return a, textinput.Blink    // Return the blink command
		}
		return a, nil // Do nothing if timer is not running
	// Add cases for other menu options (e, logs, edit, delete) here
	case "e":
		// Handle manual entry
		// This could be a new state or a function call
		// For now, just print a message
		fmt.Println("Manual entry not implemented yet.")
		return a, nil
	}
	return a, nil // No command for unhandled keys in this state
}

// when asking for stop message
func (a *app) handleKeypressStoppingTimer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c": // Allow quitting from input prompt
		return a, tea.Quit
	case "enter":
		// Stop the timer with the message from the input
		message := a.message.Value()
		if a.project != nil {
			a.project.StopTimer(message, true) // StopTimer handles empty message ok
		}
		a.project = nil            // Clear the selected project
		a.state = stateProjectList // Go back to the project list
		a.message.Blur()           // Unfocus the input
		// Potentially trigger a refresh of the project list if needed
		return a, nil
	case "esc":
		// Cancel stopping, go back to project menu
		a.state = stateProjectMenu
		a.message.Blur() // Unfocus the input
		return a, nil
	}

	// Update the text input model
	var cmd tea.Cmd
	a.message, cmd = a.message.Update(msg)
	return a, cmd
}

func (a app) View() string {
	switch a.state {
	case stateProjectList:
		return a.projects.View()

	case stateProjectMenu:
		// Render the menu for the selected project
		if a.project == nil {
			return "Error: No project selected.\nPress Esc to return to list." // Should not happen
		}

		var lines []string
		onclock, _ := a.project.OnClock()
		var headerText string
		if onclock {
			headerText = "Timer is running for project: " + *a.project.Name
		} else {
			headerText = "Choose an option for project: " + *a.project.Name
		}
		lines = append(lines, titleStyle.MarginTop(1).Render(headerText))
		lines = append(lines, "") // Add a blank line for spacing

		// Render each choice as a separate line
		for _, choice := range a.choices {
			// Skip "Start timer" if already running
			if onclock && choice[0] == "s" {
				continue
			}
			// Skip "End timer" if not running
			if !onclock && choice[0] == "p" {
				continue
			}

			// Format the choice string only if key is present
			var choiceText string
			if choice[0] != "" {
				choiceText = fmt.Sprintf("[%s] %s", choice[0], choice[1])
			} else {
				choiceText = fmt.Sprintf("    %s", choice[1]) // Indent options without keys
			}
			lines = append(lines, itemStyle.Render(choiceText))
		}

		lines = append(lines, "") // Add spacing before help
		helpText := helpStyle.Render("esc: back | q: quit")
		lines = append(lines, helpText)

		return lipgloss.JoinVertical(lipgloss.Left, lines...)

	case stateStoppingTimer:
		// Render the stop message input prompt
		var lines []string
		promptText := "Enter message for stopping timer (Project: " + *a.project.Name + ")"
		lines = append(lines, titleStyle.MarginTop(1).Render(promptText))
		lines = append(lines, "") // Blank line
		lines = append(lines, inputPromptStyle.Render(a.message.View()))
		lines = append(lines, "") // Blank line
		helpText := helpStyle.Render("enter: submit | esc: cancel | ctrl+c: quit")
		lines = append(lines, helpText)

		return lipgloss.JoinVertical(lipgloss.Left, lines...)

	default:
		return "Unknown state"
	}
}
