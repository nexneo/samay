package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nexneo/samay/data"
)

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
		a.manualEntryFocus = focusTime
		a.manualBillable = true
		return a, nil
	case "enter", "tab", "shift+tab", "up", "down":
		const manualFocusCount = 3
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

			entry := a.project.CreateEntryWithDuration(message, duration, a.manualBillable)
			err = data.Save(entry)
			if err != nil {
				a.errorMessage = fmt.Sprintf("Error saving entry: %v", err)
				return a, nil
			}

			a.state = stateProjectMenu
			a.manualTimeInput.Blur()
			a.manualMsgInput.Blur()
			a.manualEntryFocus = focusTime
			a.manualBillable = true
			return a, nil
		}

		if keypress == "enter" && a.manualEntryFocus == focusBillable {
			a.manualBillable = !a.manualBillable
			return a, textinput.Blink
		}

		var delta int
		switch keypress {
		case "tab", "enter", "down":
			delta = 1
		case "shift+tab", "up":
			delta = -1
		default:
			return a, nil
		}

		a.manualEntryFocus = manualFocus((int(a.manualEntryFocus) + delta + manualFocusCount) % manualFocusCount)

		switch a.manualEntryFocus {
		case focusTime:
			a.manualTimeInput.Focus()
			a.manualMsgInput.Blur()
		case focusMessage:
			a.manualTimeInput.Blur()
			a.manualMsgInput.Focus()
		case focusBillable:
			a.manualTimeInput.Blur()
			a.manualMsgInput.Blur()
		}
		// Always blink cursor on focus change
		return a, textinput.Blink

	default: // Handle regular character input
		switch a.manualEntryFocus {
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
