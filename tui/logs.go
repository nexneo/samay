package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/nexneo/samay/data"
)

// Helper view for log title
func (a app) logTitleView() string {
	if a.project == nil {
		return "" // No title if no project
	}
	title := fmt.Sprintf("Logs for Project: %s", *a.project.Name)
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

	var sb strings.Builder
	var day int = -1 // Initialize day to -1 to ensure the first header prints
	var total int64
	now := time.Now()
	maxEntries := 30 // Limit number of entries displayed
	if a.logShowAll {
		maxEntries = len(entries)
	}

	printHeader := func(ty *time.Time) {
		headerStr := ""
		if now.Year() == ty.Year() && now.YearDay() == ty.YearDay() {
			headerStr = fmt.Sprintf("%s%s", headerStr, "Today")
		} else {
			// Format as MM/DD or YYYY/MM/DD if different year
			if now.Year() == ty.Year() {
				headerStr = fmt.Sprintf("%s%s", headerStr, ty.Format("01/02")) // MM/DD
			} else {
				headerStr = fmt.Sprintf("%s%s", headerStr, ty.Format("2006/01/02")) // YYYY/MM/DD
			}
		}
		sb.WriteString("\n")
		sb.WriteString(logHeaderStyle.Render(headerStr))
		sb.WriteString("\n") // Newline after header
	}

	printTotal := func(totalDuration int64) {
		if totalDuration != 0 {
			// Use your existing HmFromD function if available and adapted
			// Assuming HmFromD returns a struct with a String() method
			// If HmFromD is not available, format manually:
			totalDur := time.Duration(totalDuration)
			totalStr := data.HmFromD(totalDur) // Use a helper if HmFromD not available
			// Right-align the total within a reasonable width (e.g., 18 chars like original)
			totalLine := fmt.Sprintf("%9s", totalStr)
			sb.WriteString(logTotalStyle.Render(totalLine))
			sb.WriteString("\n") // Newline after total
		}
	}
	// --- End Helper Functions ---

	// Main Title
	sb.WriteString(logTitleStyle.Render(" #  Hours    Description"))
	sb.WriteString("\n")
	sb.WriteString(logTitleStyle.Render("------------------------------------")) // Separator
	sb.WriteString("\n")

	entryCount := 0
	for i, entry := range entries {
		if entryCount >= maxEntries {
			sb.WriteString(itemStyle.Render(fmt.Sprintf("... (showing last %d entries)", maxEntries)))
			sb.WriteString("\n")
			break
		}

		ty, err := entry.EndedTime()
		if err != nil {
			// Handle error - maybe skip entry or show an error message?
			sb.WriteString(errorStyle.Render(fmt.Sprintf("  Error getting time for entry %d: %v\n", i, err)))
			continue // Skip this entry
		}

		// Check if the day has changed
		currentDay := ty.YearDay() // Use YearDay for comparison across year boundaries
		if day != currentDay {
			printTotal(total) // Print total for the previous day
			printHeader(ty)   // Print header for the new day
			day = currentDay  // Update the current day
			total = 0         // Reset total for the new day
		}

		// Add duration to total
		if entry.Duration != nil {
			total += *entry.Duration
		}

		// Format the entry line
		// Adjust description length based on available width (simple truncation)
		// Example: Max description width = total width - index width - hours width - padding
		descMaxWidth := width - 4 - 8 - 4 // Rough estimate, adjust as needed
		if descMaxWidth < 10 {
			descMaxWidth = 10 // Minimum width
		}
		desc := entry.GetContent()
		if len(desc) > descMaxWidth {
			desc = desc[:descMaxWidth-3] + "..." // Truncate with ellipsis
		}

		// Use entry.HoursMins() if it exists and returns a string
		// Otherwise format duration manually
		hoursMinsStr := data.HmFromD(time.Duration(*entry.Duration)) // Use helper

		// %2d: index (right-aligned, 2 spaces)
		// %-8s: hours/mins (left-aligned, 8 spaces) - adjust width as needed
		// %s: description
		entryLine := fmt.Sprintf("%2d %-8s %s", entryCount+1, hoursMinsStr, desc)
		sb.WriteString(logEntryStyle.Render(entryLine))
		sb.WriteString("\n") // Newline after each entry
		entryCount++
	}

	// Print the total for the last day
	printTotal(total)
	sb.WriteString("\n") // Extra newline at the end

	return sb.String()
}
