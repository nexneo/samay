package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nexneo/samay/data"
)

// Helper view for log title
func (a app) logTitleView() string {
	if a.project == nil {
		return "" // No title if no project
	}
	title := fmt.Sprintf("Logs for Project: %s", a.project.Name)
	return titleStyle.Render(title)
}

// Helper view for log help
func (a app) logHelpView() string {
	return helpStyle.Render("↑/↓/j/k: scroll | a: toggle range | esc: back | q: quit")
}

// formatProjectLogs generates the log string for the viewport
// It takes the available width to potentially handle wrapping (though basic for now)
func (a *app) formatProjectLogs(project *data.Project, width int) string {
	if project == nil {
		return errorStyle.Render("Error: No project data available.")
	}

	entries := project.Entries() // Get entries (assuming sorted descending by time)
	if len(entries) == 0 {
		return itemStyle.Render("No log entries found for this project.")
	}

	const (
		indexColWidth = 3
		hoursColWidth = 7
	)
	columnSpacer := "  "
	minContentWidth := indexColWidth + hoursColWidth + len(columnSpacer)*2 + 10
	if width < minContentWidth {
		width = minContentWidth
	}

	descMaxWidth := width - indexColWidth - hoursColWidth - len(columnSpacer)*2
	if descMaxWidth < 10 {
		descMaxWidth = 10
	}

	var sb strings.Builder
	dayKey := -1 // Initialize day to ensure the first header prints
	var dayTotal time.Duration
	now := time.Now()
	maxEntries := 30 // Limit number of entries displayed
	if a.logShowAll {
		maxEntries = len(entries)
	}

	headerLine := fmt.Sprintf("%-*s%s%-*s%s%s",
		indexColWidth,
		"#",
		columnSpacer,
		hoursColWidth,
		"Hours",
		columnSpacer,
		"Description",
	)
	separatorLine := strings.Repeat("-", len(headerLine))
	indentForHeader := strings.Repeat(" ", indexColWidth+hoursColWidth+len(columnSpacer)*2)

	sb.WriteString(logTitleStyle.Render(headerLine))
	sb.WriteString("\n")
	sb.WriteString(logTitleStyle.Render(separatorLine))
	sb.WriteString("\n")

	printHeader := func(ty *time.Time) {
		var headerStr string
		if now.Year() == ty.Year() && now.YearDay() == ty.YearDay() {
			headerStr = "Today"
		} else {
			// Format as MM/DD or YYYY/MM/DD if different year
			if now.Year() == ty.Year() {
				headerStr = ty.Format("01/02") // MM/DD
			} else {
				headerStr = ty.Format("2006/01/02") // YYYY/MM/DD
			}
		}
		sb.WriteString("\n")
		sb.WriteString(logHeaderStyle.Render(indentForHeader + headerStr))
		sb.WriteString("\n") // Newline after header
	}

	printTotal := func(totalDuration time.Duration) {
		if totalDuration == 0 {
			return
		}
		totalStr := data.HmFromD(totalDuration).String()
		totalLine := fmt.Sprintf("%-*s%s%-*s%s%s",
			indexColWidth,
			"",
			columnSpacer,
			hoursColWidth,
			totalStr,
			columnSpacer,
			"",
		)
		sb.WriteString(logTotalStyle.Render(totalLine))
		sb.WriteString("\n") // Newline after total
	}
	// --- End Helper Functions ---

	entryCount := 0
	for i, entry := range entries {
		if entryCount >= maxEntries {
			sb.WriteString(itemStyle.Render(fmt.Sprintf("... (showing last %d entries)", maxEntries)))
			sb.WriteString("\n")
			break
		}

		ty, err := entry.EndedTime()
		if err != nil {
			sb.WriteString(errorStyle.Render(fmt.Sprintf("Error getting time for entry %d: %v\n", i, err)))
			continue // Skip this entry
		}

		if ty == nil {
			continue
		}

		// Check if the day has changed
		currentDayKey := ty.Year()*1000 + ty.YearDay()
		if dayKey != currentDayKey {
			printTotal(dayTotal) // Print total for the previous day
			printHeader(ty)      // Print header for the new day
			dayKey = currentDayKey
			dayTotal = 0 // Reset total for the new day
		}

		// Add duration to total
		duration := time.Duration(entry.GetDuration())
		if duration > 0 {
			dayTotal += duration
		}
		durationStr := "--:--"
		if duration > 0 {
			durationStr = data.HmFromD(duration).String()
		}

		// Format description with rune-aware truncation
		desc := entry.GetContent()
		descRunes := []rune(desc)
		if len(descRunes) > descMaxWidth {
			desc = string(descRunes[:descMaxWidth-3]) + "..."
		}

		entryLine := fmt.Sprintf("%-*d%s%-*s%s%s",
			indexColWidth,
			entryCount+1,
			columnSpacer,
			hoursColWidth,
			durationStr,
			columnSpacer,
			desc,
		)
		sb.WriteString(logEntryStyle.Render(entryLine))
		sb.WriteString("\n") // Newline after each entry
		entryCount++
	}

	// Print the total for the last day
	printTotal(dayTotal)
	sb.WriteString("\n") // Extra newline at the end

	return sb.String()
}

// when showing logs
func (a *app) handleKeypressShowLogs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "esc":
		a.state = stateProjectMenu
		a.errorMessage = "" // Clear log-related errors
		return a, nil
	case "a":
		a.logShowAll = !a.logShowAll
		a.ProjectLogUI()
		return a, nil
	}

	// Handle viewport scrolling (up/down arrows, pgup/pgdown, j/k)
	a.logViewport, cmd = a.logViewport.Update(msg)
	return a, cmd
}
