package tui

import (
	"fmt"
	"time" // Import the time package

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
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).PaddingLeft(2) // Style for error messages
)

// Define different states for the application
type state int

const (
	stateProjectList   state = iota // Showing the list of projects
	stateProjectMenu                // Showing the menu for a selected project
	stateStoppingTimer              // Asking for a message before stopping timer
	stateManualEntry                // Asking for time and message for manual entry
)

// Define focus states for manual entry
type manualFocus int

const (
	focusTime manualFocus = iota
	focusMessage
)

type app struct {
	project          *data.Project
	projects         list.Model
	choices          [][2]string
	state            state
	stopMessageInput textinput.Model // Renamed for clarity
	manualTimeInput  textinput.Model // Input for manual entry time
	manualMsgInput   textinput.Model // Input for manual entry message
	manualEntryFocus manualFocus     // Which input is focused in manual entry
	width            int             // store window width
	errorMessage     string          // To display temporary errors
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

	// text input model for stopping timer
	stopTI := textinput.New()
	stopTI.Placeholder = "Enter stop message (optional, press Enter to submit, Esc to cancel)"
	stopTI.CharLimit = 156
	stopTI.Width = 50 // Adjust width as needed

	// text input models for manual entry
	manualTimeTI := textinput.New()
	manualTimeTI.Placeholder = "e.g., 1h30m, 45m"
	manualTimeTI.CharLimit = 20
	manualTimeTI.Width = 20

	manualMsgTI := textinput.New()
	manualMsgTI.Placeholder = "Description of the work done"
	manualMsgTI.CharLimit = 156
	manualMsgTI.Width = 50

	initialState := stateProjectList
	if currentProject != nil {
		// show project menu if a timer is running, otherwise show project list
		initialState = stateProjectMenu
	}

	return &app{
		project:          currentProject,
		projects:         l,
		state:            initialState,
		stopMessageInput: stopTI, // Use the renamed field
		manualTimeInput:  manualTimeTI,
		manualMsgInput:   manualMsgTI,
		manualEntryFocus: focusTime, // Start focus on time input
		choices: [][2]string{
			{"s", "Start timer"},
			{"p", "End timer"},
			{"e", "Enter manually"},
			{"", "Show logs"},      // Placeholder
			{"", "Edit project"},   // Placeholder
			{"", "Delete project"}, // Placeholder
		},
	}
}

func (a app) Init() tea.Cmd {
	return nil // no initial command for now
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd // Slice to hold commands

	// Clear error message on any key press or resize
	a.errorMessage = ""

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width // Store width
		a.projects.SetWidth(msg.Width)
		// Adjust input widths dynamically if desired
		// a.stopMessageInput.Width = msg.Width - 10
		// a.manualTimeInput.Width = 20
		// a.manualMsgInput.Width = msg.Width - 30
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
		case stateManualEntry:
			m, c := a.handleKeypressManualEntry(msg)
			return m, c
		}
	}

	// If not handled by state-specific keypress logic, update the relevant model
	switch a.state {
	case stateProjectList:
		a.projects, cmd = a.projects.Update(msg)
		cmds = append(cmds, cmd)
	case stateStoppingTimer:
		a.stopMessageInput, cmd = a.stopMessageInput.Update(msg)
		cmds = append(cmds, cmd)
	case stateManualEntry:
		// Update the currently focused input field
		if a.manualEntryFocus == focusTime {
			a.manualTimeInput, cmd = a.manualTimeInput.Update(msg)
		} else {
			a.manualMsgInput, cmd = a.manualMsgInput.Update(msg)
		}
		cmds = append(cmds, cmd)
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
				a.state = stateProjectMenu // Switch to project menu
				// Potentially load project details or check timer status here
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
	if a.project == nil { // Safety check
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
	case "s": // Start Timer
		if !onclock {
			err := a.project.StartTimer()
			if err != nil {
				a.errorMessage = fmt.Sprintf("Error starting timer: %v", err)
			}
			// No state change needed, view will update based on onclock status
		}
		return a, nil
	case "p": // End Timer (Prepare)
		if onclock {
			a.state = stateStoppingTimer    // Switch to stopping state
			a.stopMessageInput.Focus()      // Focus the text input
			a.stopMessageInput.SetValue("") // Clear previous message
			return a, textinput.Blink       // Return the blink command
		}
		return a, nil // Do nothing if timer is not running
	case "e": // Enter Manually (Prepare)
		a.state = stateManualEntry     // Switch to manual entry state
		a.manualEntryFocus = focusTime // Start with time input focused
		a.manualTimeInput.Focus()
		a.manualMsgInput.Blur()
		a.manualTimeInput.SetValue("") // Clear previous values
		a.manualMsgInput.SetValue("")
		return a, textinput.Blink // Blink the time input cursor
		// Add cases for other menu options (logs, edit, delete) here
	}
	return a, nil // No command for unhandled keys in this state
}

// when asking for stop message
func (a *app) handleKeypressStoppingTimer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c":
		return a, tea.Quit
	case "enter":
		// Stop the timer with the message from the input
		message := a.stopMessageInput.Value()
		if a.project != nil {
			err := a.project.StopTimer(message, true) // StopTimer handles empty message ok
			if err != nil {
				// If stopping fails, show error and return to menu
				a.errorMessage = fmt.Sprintf("Error stopping timer: %v", err)
				a.state = stateProjectMenu
				a.stopMessageInput.Blur()
				return a, nil
			}
		}
		// Success: Go back to the project list and clear project
		a.project = nil
		a.state = stateProjectList
		a.stopMessageInput.Blur()
		// Potentially trigger a refresh of the project list if needed
		return a, nil
	case "esc":
		// Cancel stopping, go back to project menu
		a.state = stateProjectMenu
		a.stopMessageInput.Blur() // Unfocus the input
		return a, nil
	}

	// Update the text input model
	var cmd tea.Cmd
	a.stopMessageInput, cmd = a.stopMessageInput.Update(msg)
	return a, cmd
}

// when asking for manual entry details
func (a *app) handleKeypressManualEntry(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch keypress := msg.String(); keypress {
	case "ctrl+c":
		return a, tea.Quit
	case "esc":
		// Cancel manual entry, go back to project menu
		a.state = stateProjectMenu
		a.manualTimeInput.Blur()
		a.manualMsgInput.Blur()
		return a, nil
	case "enter", "tab", "shift+tab", "up", "down": // Handle focus switching and submission
		if keypress == "enter" && a.manualEntryFocus == focusMessage {
			// --- Try to save the manual entry ---
			durationStr := a.manualTimeInput.Value()
			message := a.manualMsgInput.Value()

			if durationStr == "" {
				a.errorMessage = "Error: Duration cannot be empty."
				a.manualEntryFocus = focusTime // Refocus time input
				a.manualTimeInput.Focus()
				a.manualMsgInput.Blur()
				return a, textinput.Blink
			}
			if message == "" {
				a.errorMessage = "Error: Message cannot be empty."
				// Keep focus on message input
				return a, textinput.Blink
			}

			// Parse the duration string (e.g., "1h30m", "45m")
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				a.errorMessage = fmt.Sprintf("Error parsing duration: %v", err)
				a.manualEntryFocus = focusTime // Refocus time input on error
				a.manualTimeInput.Focus()
				a.manualMsgInput.Blur()
				return a, textinput.Blink
			}

			if a.project == nil {
				a.errorMessage = "Error: No project selected (internal error)."
				a.state = stateProjectList // Go back to list if project is lost
				return a, nil
			}

			// Create and save the entry (assuming billable is true for now)
			billable := true // Or get this from project settings/another input
			entry := a.project.CreateEntryWithDuration(message, duration, billable)
			err = data.Save(entry) // Use the Save function from data package
			if err != nil {
				a.errorMessage = fmt.Sprintf("Error saving entry: %v", err)
				// Stay in manual entry state to allow correction or cancellation
				return a, nil
			}

			// Success: Go back to project menu
			a.state = stateProjectMenu
			a.manualTimeInput.Blur()
			a.manualMsgInput.Blur()
			// Optionally display a success message briefly?
			return a, nil

		} else {
			// --- Switch focus between inputs ---
			if keypress == "tab" || keypress == "enter" || keypress == "down" {
				a.manualEntryFocus = (a.manualEntryFocus + 1) % 2 // Cycle focus forward
			} else if keypress == "shift+tab" || keypress == "up" {
				a.manualEntryFocus = (a.manualEntryFocus + 2 - 1) % 2 // Cycle focus backward (modulo arithmetic)
			}

			if a.manualEntryFocus == focusTime {
				a.manualTimeInput.Focus()
				a.manualMsgInput.Blur()
				cmd = textinput.Blink
			} else {
				a.manualTimeInput.Blur()
				a.manualMsgInput.Focus()
				cmd = textinput.Blink
			}
			return a, cmd
		}

	default: // Handle regular character input for the focused field
		if a.manualEntryFocus == focusTime {
			a.manualTimeInput, cmd = a.manualTimeInput.Update(msg)
		} else {
			a.manualMsgInput, cmd = a.manualMsgInput.Update(msg)
		}
		return a, cmd
	}
}

func (a app) View() string {
	var viewContent string

	switch a.state {
	case stateProjectList:
		viewContent = a.projects.View()

	case stateProjectMenu:
		if a.project == nil {
			viewContent = "Error: No project selected.\nPress Esc to return to list." // Should not happen
		} else {
			var lines []string
			onclock, _ := a.project.OnClock()
			var headerText string
			if onclock {
				headerText = "Timer is running for project: " + *a.project.Name
			} else {
				headerText = "Choose an option for project: " + *a.project.Name
			}
			lines = append(lines, titleStyle.MarginTop(1).Render(headerText))
			lines = append(lines, "") // Spacing

			// Render choices
			for _, choice := range a.choices {
				// Skip "Start timer" if already running
				if onclock && choice[0] == "s" {
					continue
				}
				// Skip "End timer" if not running
				if !onclock && choice[0] == "p" {
					continue
				}

				var choiceText string
				if choice[0] != "" {
					choiceText = fmt.Sprintf("[%s] %s", choice[0], choice[1])
				} else {
					choiceText = fmt.Sprintf("    %s", choice[1]) // Indent options without keys
				}
				lines = append(lines, itemStyle.Render(choiceText))
			}

			lines = append(lines, "") // Spacing before help
			helpText := helpStyle.Render("esc: back | q: quit")
			lines = append(lines, helpText)

			viewContent = lipgloss.JoinVertical(lipgloss.Left, lines...)
		}

	case stateStoppingTimer:
		var lines []string
		promptText := "Enter message for stopping timer (Project: " + *a.project.Name + ")"
		lines = append(lines, titleStyle.MarginTop(1).Render(promptText))
		lines = append(lines, "") // Blank line
		lines = append(lines, inputPromptStyle.Render(a.stopMessageInput.View()))
		lines = append(lines, "") // Blank line
		helpText := helpStyle.Render("enter: submit | esc: cancel | ctrl+c: quit")
		lines = append(lines, helpText)
		viewContent = lipgloss.JoinVertical(lipgloss.Left, lines...)

	case stateManualEntry:
		var lines []string
		promptText := "Manually enter time for project: " + *a.project.Name
		lines = append(lines, titleStyle.MarginTop(1).Render(promptText))
		lines = append(lines, "") // Blank line

		// Render Time Input
		lines = append(lines, inputPromptStyle.Render("Duration (e.g., 1h30m):"))
		lines = append(lines, itemStyle.Render(a.manualTimeInput.View()))
		lines = append(lines, "") // Blank line

		// Render Message Input
		lines = append(lines, inputPromptStyle.Render("Message:"))
		lines = append(lines, itemStyle.Render(a.manualMsgInput.View()))
		lines = append(lines, "") // Blank line

		helpText := helpStyle.Render("enter: next/submit | tab/↑/↓: switch | esc: cancel | ctrl+c: quit")
		lines = append(lines, helpText)
		viewContent = lipgloss.JoinVertical(lipgloss.Left, lines...)

	default:
		viewContent = "Unknown state"
	}

	// Append error message if any
	if a.errorMessage != "" {
		// Add a blank line before the error if content exists
		if viewContent != "" {
			viewContent += "\n\n"
		}
		viewContent += errorStyle.Render(a.errorMessage)
	}

	return viewContent
}
