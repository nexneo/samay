package tui

import (
	"fmt"

	"time" // Import the time package

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport" // Import viewport for scrolling logs
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nexneo/samay/data" // Assuming util.Color exists and HmFromD
	"github.com/samber/lo"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	inputPromptStyle  = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("240")) // Style for input prompt
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).PaddingLeft(2) // Style for error messages
	logHeaderStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("37")).Bold(true)      // Style for log date headers
	logTotalStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Bold(true)      // Style for log totals
	logEntryStyle     = lipgloss.NewStyle()                                                  // Style for individual log entries
	logTitleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)      // Style for the main log title
)

// Define different states for the application
type state int

const (
	stateProjectList   state = iota // Showing the list of projects
	stateProjectMenu                // Showing the menu for a selected project
	stateStoppingTimer              // Asking for a message before stopping timer
	stateManualEntry                // Asking for time and message for manual entry
	stateShowLogs                   // Displaying project logs
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
	logViewport      viewport.Model  // Viewport for scrolling logs
	width            int             // store window width
	height           int             // store window height
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

	// Viewport for logs
	vp := viewport.New(defaultWidth, 20) // Initial size, will be updated
	vp.SetContent("Loading logs...")     // Placeholder content

	initialState := stateProjectList
	if currentProject != nil {
		// show project menu if a timer is running, otherwise show project list
		initialState = stateProjectMenu
	}

	return &app{
		project:          currentProject,
		projects:         l,
		state:            initialState,
		stopMessageInput: stopTI,
		manualTimeInput:  manualTimeTI,
		manualMsgInput:   manualMsgTI,
		manualEntryFocus: focusTime,
		logViewport:      vp,
		choices: [][2]string{
			{"s", "Start timer"},
			{"p", "End timer"},
			{"e", "Enter manually"},
			{"l", "Show logs"},     // Added 'l' keybind
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

	// Clear error message on any key press or resize, unless we are showing logs
	// where the error might be relevant to the log fetching itself.
	if a.state != stateShowLogs {
		a.errorMessage = ""
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width   // Store width
		a.height = msg.Height // Store height
		a.projects.SetWidth(msg.Width)
		// Update viewport size, leave some space for title and help
		headerHeight := lipgloss.Height(a.logTitleView())
		footerHeight := lipgloss.Height(a.logHelpView())
		a.logViewport.Width = msg.Width
		a.logViewport.Height = msg.Height - headerHeight - footerHeight
		// Adjust input widths dynamically if desired
		// a.stopMessageInput.Width = msg.Width - 10
		// a.manualMsgInput.Width = msg.Width - 30
		// Re-render logs if we are in that state, as width might affect wrapping
		if a.state == stateShowLogs && a.project != nil {
			a.logViewport.SetContent(a.formatProjectLogs(a.project, a.logViewport.Width)) // Pass width
		}
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
		case stateShowLogs:
			m, c := a.handleKeypressShowLogs(msg)
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
		if a.manualEntryFocus == focusTime {
			a.manualTimeInput, cmd = a.manualTimeInput.Update(msg)
		} else {
			a.manualMsgInput, cmd = a.manualMsgInput.Update(msg)
		}
		cmds = append(cmds, cmd)
	case stateShowLogs:
		// Update the viewport for scrolling
		a.logViewport, cmd = a.logViewport.Update(msg)
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
			// Find the project by name from the original list
			foundProject := lo.Filter(data.DB.Projects(), func(p *data.Project, _ int) bool {
				return *p.Name == string(i)
			})
			if len(foundProject) > 0 {
				a.project = foundProject[0] // Get the actual project object
				a.state = stateProjectMenu  // Switch to project menu
			} else {
				// This case should ideally not happen if the list is sourced correctly
				a.errorMessage = fmt.Sprintf("Error: Could not find project '%s'", string(i))
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
		}
		return a, nil
	case "p": // End Timer (Prepare)
		if onclock {
			a.state = stateStoppingTimer
			a.stopMessageInput.Focus()
			a.stopMessageInput.SetValue("")
			return a, textinput.Blink
		}
		return a, nil
	case "e": // Enter Manually (Prepare)
		a.state = stateManualEntry
		a.manualEntryFocus = focusTime
		a.manualTimeInput.Focus()
		a.manualMsgInput.Blur()
		a.manualTimeInput.SetValue("")
		a.manualMsgInput.SetValue("")
		return a, textinput.Blink
	case "l": // Show Logs
		a.state = stateShowLogs
		// Format logs and set viewport content
		a.logViewport.SetContent(a.formatProjectLogs(a.project, a.logViewport.Width)) // Pass width
		a.logViewport.GotoTop()                                                       // Scroll to top initially
		a.errorMessage = ""                                                           // Clear previous errors
		return a, nil
		// Add cases for other menu options (edit, delete) here
	}
	return a, nil // No command for unhandled keys in this state
}

// when asking for stop message
func (a *app) handleKeypressStoppingTimer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c":
		return a, tea.Quit
	case "enter":
		message := a.stopMessageInput.Value()
		if a.project != nil {
			err := a.project.StopTimer(message, true)
			if err != nil {
				a.errorMessage = fmt.Sprintf("Error stopping timer: %v", err)
				a.state = stateProjectMenu
				a.stopMessageInput.Blur()
				return a, nil
			}
		}
		a.project = nil
		a.state = stateProjectList
		a.stopMessageInput.Blur()
		return a, nil
	case "esc":
		a.state = stateProjectMenu
		a.stopMessageInput.Blur()
		return a, nil
	}

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
		a.state = stateProjectMenu
		a.manualTimeInput.Blur()
		a.manualMsgInput.Blur()
		return a, nil
	case "enter", "tab", "shift+tab", "up", "down":
		if keypress == "enter" && a.manualEntryFocus == focusMessage {
			durationStr := a.manualTimeInput.Value()
			message := a.manualMsgInput.Value()

			if durationStr == "" {
				a.errorMessage = "Error: Duration cannot be empty."
				a.manualEntryFocus = focusTime
				a.manualTimeInput.Focus()
				a.manualMsgInput.Blur()
				return a, textinput.Blink
			}
			if message == "" {
				a.errorMessage = "Error: Message cannot be empty."
				return a, textinput.Blink
			}

			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				a.errorMessage = fmt.Sprintf("Error parsing duration: %v", err)
				a.manualEntryFocus = focusTime
				a.manualTimeInput.Focus()
				a.manualMsgInput.Blur()
				return a, textinput.Blink
			}

			if a.project == nil {
				a.errorMessage = "Error: No project selected (internal error)."
				a.state = stateProjectList
				return a, nil
			}

			billable := true // Assume billable
			entry := a.project.CreateEntryWithDuration(message, duration, billable)
			err = data.Save(entry)
			if err != nil {
				a.errorMessage = fmt.Sprintf("Error saving entry: %v", err)
				return a, nil
			}

			a.state = stateProjectMenu
			a.manualTimeInput.Blur()
			a.manualMsgInput.Blur()
			return a, nil

		} else {
			// Switch focus
			if keypress == "tab" || keypress == "enter" || keypress == "down" {
				a.manualEntryFocus = (a.manualEntryFocus + 1) % 2
			} else if keypress == "shift+tab" || keypress == "up" {
				a.manualEntryFocus = (a.manualEntryFocus + 2 - 1) % 2
			}

			if a.manualEntryFocus == focusTime {
				a.manualTimeInput.Focus()
				a.manualMsgInput.Blur()
			} else {
				a.manualTimeInput.Blur()
				a.manualMsgInput.Focus()
			}
			// Always blink cursor on focus change
			return a, textinput.Blink
		}

	default: // Handle regular character input
		if a.manualEntryFocus == focusTime {
			a.manualTimeInput, cmd = a.manualTimeInput.Update(msg)
		} else {
			a.manualMsgInput, cmd = a.manualMsgInput.Update(msg)
		}
		return a, cmd
	}
}

// when showing logs
func (a *app) handleKeypressShowLogs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch keypress := msg.String(); keypress {
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "q": // Go back to project menu
		a.state = stateProjectMenu
		a.errorMessage = "" // Clear log-related errors
		return a, nil
	}

	// Handle viewport scrolling (up/down arrows, pgup/pgdown, j/k)
	a.logViewport, cmd = a.logViewport.Update(msg)
	return a, cmd
}

func (a app) View() string {
	var viewContent string

	switch a.state {
	case stateProjectList:
		viewContent = a.projects.View()

	case stateProjectMenu:
		if a.project == nil {
			viewContent = errorStyle.Render("Error: No project selected.\nPress Esc to return to list.")
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
				if onclock && choice[0] == "s" {
					continue
				}
				if !onclock && choice[0] == "p" {
					continue
				}

				var choiceText string
				if choice[0] != "" {
					choiceText = fmt.Sprintf("[%s] %s", choice[0], choice[1])
				} else {
					choiceText = fmt.Sprintf("    %s", choice[1])
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
		lines = append(lines, "")
		lines = append(lines, inputPromptStyle.Render(a.stopMessageInput.View()))
		lines = append(lines, "")
		helpText := helpStyle.Render("enter: submit | esc: cancel | ctrl+c: quit")
		lines = append(lines, helpText)
		viewContent = lipgloss.JoinVertical(lipgloss.Left, lines...)

	case stateManualEntry:
		var lines []string
		promptText := "Manually enter time for project: " + *a.project.Name
		lines = append(lines, titleStyle.MarginTop(1).Render(promptText))
		lines = append(lines, "")
		lines = append(lines, inputPromptStyle.Render("Duration (e.g., 1h30m):"))
		lines = append(lines, itemStyle.Render(a.manualTimeInput.View()))
		lines = append(lines, "")
		lines = append(lines, inputPromptStyle.Render("Message:"))
		lines = append(lines, itemStyle.Render(a.manualMsgInput.View()))
		lines = append(lines, "")
		helpText := helpStyle.Render("enter: next/submit | tab/↑/↓: switch | esc: cancel | ctrl+c: quit")
		lines = append(lines, helpText)
		viewContent = lipgloss.JoinVertical(lipgloss.Left, lines...)

	case stateShowLogs:
		// Combine title, viewport, and help view
		titleView := a.logTitleView()
		helpView := a.logHelpView()
		// Use JoinVertical for proper layout respecting viewport height
		viewContent = lipgloss.JoinVertical(lipgloss.Left,
			titleView,
			a.logViewport.View(), // Render the viewport content
			helpView,
		)

	default:
		viewContent = errorStyle.Render("Unknown state")
	}

	// Append error message if any (and not empty)
	// Ensure it doesn't overwrite the entire log view if there's a log error
	if a.errorMessage != "" {
		errorMsgRendered := errorStyle.Render(a.errorMessage)
		// If we are showing logs, append the error below the help text
		if a.state == stateShowLogs {
			viewContent = lipgloss.JoinVertical(lipgloss.Left, viewContent, errorMsgRendered)
		} else {
			// For other states, append with spacing
			viewContent = lipgloss.JoinVertical(lipgloss.Left, viewContent, "", errorMsgRendered)
		}

	}

	return viewContent
}
