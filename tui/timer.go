package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// when asking for stop message
func (a *app) handleKeypressStoppingTimer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeySpace && a.stopEntryFocus == focusStopBillable {
		a.stopBillable = !a.stopBillable
		return a, textinput.Blink
	}

	switch keypress := msg.String(); keypress {
	case "ctrl+c":
		return a, tea.Quit
	case "enter", "tab", "shift+tab", "up", "down":
		const stopFocusCount = 2
		if keypress == "enter" && a.stopEntryFocus == focusStopMessage {
			message := a.stopMessageInput.Value()
			if a.project != nil {
				if err := a.project.StopTimer(message, a.stopBillable); err != nil {
					a.errorMessage = fmt.Sprintf("Error stopping timer: %v", err)
				} else {
					a.refreshEntryList()
				}
			}
			a.state = stateProjectMenu
			a.updateProjectSelectionFromList()
			a.stopMessageInput.SetValue("")
			a.stopMessageInput.Blur()
			a.stopBillable = true
			a.stopEntryFocus = focusStopMessage
			return a, tea.ClearScreen
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

		a.stopEntryFocus = stopFocus((int(a.stopEntryFocus) + delta + stopFocusCount) % stopFocusCount)
		if a.stopEntryFocus == focusStopMessage {
			a.stopMessageInput.Focus()
		} else {
			a.stopMessageInput.Blur()
		}
		return a, textinput.Blink
	case "esc":
		a.state = stateProjectMenu
		a.stopMessageInput.SetValue("")
		a.stopMessageInput.Blur()
		a.stopBillable = true
		a.stopEntryFocus = focusStopMessage
		return a, tea.ClearScreen
	}

	if a.stopEntryFocus == focusStopMessage {
		var cmd tea.Cmd
		a.stopMessageInput, cmd = a.stopMessageInput.Update(msg)
		return a, cmd
	}

	return a, nil
}
