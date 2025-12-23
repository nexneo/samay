package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// when asking for manual entry details
func (a *app) handleKeypressManualEntry(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if msg.Type == tea.KeySpace && a.manualEntryFocus == focusBillable {
		a.manualBillable = !a.manualBillable
		return a, textinput.Blink
	}

	switch keypress := msg.String(); keypress {
	case "ctrl+c":
		return a, tea.Quit
	case "esc":
		a.state = stateProjectMenu
		a.manualDateInput.Blur()
		a.manualTimeInput.Blur()
		a.manualMsgInput.Blur()
		a.manualEntryFocus = focusDate
		a.manualBillable = true
		a.manualDateInput.Reset()
		a.manualTimeInput.Reset()
		a.manualMsgInput.Reset()
		return a, nil
	case "enter":
		dateStr := a.manualDateInput.Value()
		durationStr := a.manualTimeInput.Value()
		message := a.manualMsgInput.Value()

		// Parse date, default to today
		entryDate := time.Now()
		if dateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				a.errorMessage = fmt.Sprintf("Error parsing date: %v", err)
				a.manualEntryFocus = focusDate
				a.manualDateInput.Focus()
				a.manualTimeInput.Blur()
				a.manualMsgInput.Blur()
				return a, textinput.Blink
			}
			entryDate = parsedDate
		}

		if durationStr == "" {
			a.errorMessage = "Error: Duration cannot be empty."
			a.manualEntryFocus = focusTime
			a.manualDateInput.Blur()
			a.manualTimeInput.Focus()
			a.manualMsgInput.Blur()
			return a, textinput.Blink
		}

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			a.errorMessage = fmt.Sprintf("Error parsing duration: %v", err)
			a.manualEntryFocus = focusTime
			a.manualDateInput.Blur()
			a.manualTimeInput.Focus()
			a.manualMsgInput.Blur()
			return a, textinput.Blink
		}

		if message == "" {
			a.errorMessage = "Error: Message cannot be empty."
			a.manualEntryFocus = focusMessage
			a.manualDateInput.Blur()
			a.manualTimeInput.Blur()
			a.manualMsgInput.Focus()
			return a, textinput.Blink
		}

		if a.project == nil {
			a.errorMessage = "Error: No project selected (internal error)."
			a.state = stateProjectList
			return a, nil
		}

		if _, err := a.project.CreateEntryWithDurationAndDate(message, duration, entryDate, a.manualBillable); err != nil {
			a.errorMessage = fmt.Sprintf("Error saving entry: %v", err)
			return a, nil
		}

		a.refreshEntryList()
		a.state = stateProjectMenu
		a.manualDateInput.Blur()
		a.manualTimeInput.Blur()
		a.manualMsgInput.Blur()
		a.manualEntryFocus = focusDate
		a.manualBillable = true
		a.manualDateInput.Reset()
		a.manualTimeInput.Reset()
		a.manualMsgInput.Reset()
		return a, nil
	case "tab", "shift+tab", "up", "down":
		const manualFocusCount = 4
		var delta int
		switch keypress {
		case "tab", "down":
			delta = 1
		case "shift+tab", "up":
			delta = -1
		default:
			return a, nil
		}

		a.manualEntryFocus = manualFocus((int(a.manualEntryFocus) + delta + manualFocusCount) % manualFocusCount)

		switch a.manualEntryFocus {
		case focusDate:
			a.manualDateInput.Focus()
			a.manualTimeInput.Blur()
			a.manualMsgInput.Blur()
		case focusTime:
			a.manualDateInput.Blur()
			a.manualTimeInput.Focus()
			a.manualMsgInput.Blur()
		case focusMessage:
			a.manualDateInput.Blur()
			a.manualTimeInput.Blur()
			a.manualMsgInput.Focus()
		case focusBillable:
			a.manualDateInput.Blur()
			a.manualTimeInput.Blur()
			a.manualMsgInput.Blur()
		}
		// Always blink cursor on focus change
		return a, textinput.Blink

	default: // Handle regular character input
		switch a.manualEntryFocus {
		case focusDate:
			a.manualDateInput, cmd = a.manualDateInput.Update(msg)
			return a, cmd
		case focusTime:
			a.manualTimeInput, cmd = a.manualTimeInput.Update(msg)
			return a, cmd
		case focusMessage:
			a.manualMsgInput, cmd = a.manualMsgInput.Update(msg)
			return a, cmd
		}
		return a, nil
	}
}
