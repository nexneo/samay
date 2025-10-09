package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *app) handleKeypressReportView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "esc":
		a.state = a.previousState
		return a, nil
	case "left", "h":
		a.adjustReportMonth(-1)
		a.ReportViewUI()
		return a, nil
	case "right", "l":
		a.adjustReportMonth(1)
		a.ReportViewUI()
		return a, nil
	case "r":
		now := time.Now()
		a.reportMonth = now.Month()
		a.reportYear = now.Year()
		a.ReportViewUI()
		return a, nil
	case "enter":
		return a, nil
	}

	var cmd tea.Cmd
	a.reportViewport, cmd = a.reportViewport.Update(msg)
	return a, cmd
}
