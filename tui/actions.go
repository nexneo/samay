package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/nexneo/samay/data"
	"google.golang.org/protobuf/proto"
)

// RemoveEntryUI retains CLI `rm` entry deletion capability within the TUI.
func (a *app) RemoveEntryUI() {
	defer func() {
		a.confirmAction = confirmNone
		a.confirmEntry = nil
		a.confirmMessage = ""
	}()

	if a.project == nil {
		a.errorMessage = "No project selected."
		a.state = stateProjectList
		return
	}

	entry := a.confirmEntry
	if entry == nil {
		entry = a.selectedEntry
	}
	if entry == nil {
		a.errorMessage = "No entry chosen."
		a.state = stateEntryList
		return
	}

	entry.Project = a.project
	if err := data.Destroy(entry); err != nil {
		a.errorMessage = fmt.Sprintf("Error deleting entry: %v", err)
		a.state = stateProjectMenu
		return
	}

	a.selectedEntry = nil
	a.entrySelectionMode = entryActionView
	a.refreshEntryList()
	if a.state == stateShowLogs {
		a.ProjectLogUI()
	}
	if len(a.entries.Items()) == 0 {
		a.state = stateProjectMenu
	} else {
		a.state = stateEntryList
	}
	// Provide soft confirmation to the user
	a.errorMessage = "Entry deleted"
}

// RemoveProjectUI retains CLI project removal capability within the TUI.
func (a *app) RemoveProjectUI() {
	project := a.confirmProject
	if project == nil {
		project = a.project
	}
	if project == nil {
		a.errorMessage = "No project selected."
		return
	}

	if err := data.Destroy(project); err != nil {
		a.errorMessage = fmt.Sprintf("Error deleting project: %v", err)
		a.state = stateProjectMenu
		return
	}

	if a.project != nil && a.project.GetName() == project.GetName() {
		a.project = nil
	}
	a.refreshProjectList()
	a.state = stateProjectList
	a.errorMessage = fmt.Sprintf("Project '%s' deleted", project.GetName())
	a.confirmProject = nil
	a.confirmAction = confirmNone
	a.confirmMessage = ""
}

// MoveEntryUI retains CLI entry move capability within the TUI.
func (a *app) MoveEntryUI() {
	if a.project == nil {
		a.errorMessage = "No source project selected."
		return
	}
	if a.selectedEntry == nil {
		a.errorMessage = "No entry selected to move."
		a.state = stateEntryList
		return
	}
	if a.moveTargetProject == nil {
		a.errorMessage = "No target project selected."
		return
	}

	source := a.project
	entry := a.selectedEntry

	entry.Project = a.moveTargetProject
	if err := data.Save(entry); err != nil {
		a.errorMessage = fmt.Sprintf("Error saving entry in destination: %v", err)
		entry.Project = source
		return
	}

	entry.Project = source
	if err := data.Destroy(entry); err != nil {
		a.errorMessage = fmt.Sprintf("Entry moved but failed to remove from source: %v", err)
		entry.Project = a.moveTargetProject
		return
	}

	entry.Project = a.moveTargetProject
	a.errorMessage = fmt.Sprintf("Entry moved to '%s'", a.moveTargetProject.GetName())
	a.moveTargetProject = nil
	a.selectedEntry = nil
	a.entrySelectionMode = entryActionView
	a.refreshEntryList()
	if len(a.entries.Items()) == 0 {
		a.state = stateProjectMenu
	} else {
		a.state = stateEntryList
	}
}

// MoveProjectUI retains CLI project migration capability within the TUI.
func (a *app) MoveProjectUI() {
	if a.project == nil {
		a.errorMessage = "No project selected."
		a.state = stateProjectList
		return
	}

	newName := strings.TrimSpace(a.renameInput.Value())
	if newName == "" {
		a.errorMessage = "Project name cannot be empty."
		return
	}

	if strings.EqualFold(newName, a.project.GetName()) {
		a.renameInput.Blur()
		a.state = stateProjectMenu
		return
	}

	// Prevent name collisions.
	for _, p := range data.DB.Projects() {
		if p.GetName() == newName {
			a.errorMessage = "A project with that name already exists."
			return
		}
	}

	oldPath := data.DB.ProjectDirPath(a.project)
	oldName := a.project.GetName()

	// Apply the new name so helper methods compute the new location.
	a.project.Name = proto.String(newName)
	newPath := data.DB.ProjectDirPath(a.project)
	if err := os.Rename(oldPath, newPath); err != nil {
		a.project.Name = proto.String(oldName)
		a.errorMessage = fmt.Sprintf("Error renaming project directory: %v", err)
		return
	}

	a.project.Sha = proto.String(a.project.GetShaFromName())
	if err := data.Update(a.project); err != nil {
		a.errorMessage = fmt.Sprintf("Error persisting project rename: %v", err)
	}

	a.renameInput.Blur()
	a.refreshProjectList()
	a.errorMessage = fmt.Sprintf("Project renamed to '%s'", newName)
	a.state = stateProjectMenu
}

// ShowEntryDetailUI retains CLI entry inspection capability within the TUI.
func (a *app) ShowEntryDetailUI() {
	if a.selectedEntry == nil {
		entry := entryFromListItem(a.entries.SelectedItem())
		if entry == nil {
			a.errorMessage = "Select an entry first."
			return
		}
		a.selectedEntry = entry
	}
	if a.project != nil {
		a.selectedEntry.Project = a.project
	}
	a.state = stateEntryDetail
}

// ReportViewUI retains CLI reporting capability within the TUI.
func (a *app) ReportViewUI() {
	now := time.Now()
	if a.reportMonth == 0 {
		a.reportMonth = now.Month()
	}
	if a.reportYear == 0 {
		a.reportYear = now.Year()
	}
	start := time.Date(a.reportYear, a.reportMonth, 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)

	type reportRow struct {
		name        string
		total       time.Duration
		billable    time.Duration
		entries     int
		isOnClock   bool
		clockAmount time.Duration
	}

	rows := make([]reportRow, 0)
	var overall, overallBillable time.Duration

	for _, project := range data.DB.Projects() {
		var row reportRow
		row.name = project.GetName()
		for _, entry := range project.Entries() {
			ended, err := entry.EndedTime()
			if err != nil || ended == nil {
				continue
			}
			if ended.Before(start) || !ended.Before(end) {
				continue
			}
			dur := time.Duration(entry.GetDuration())
			row.total += dur
			if entry.GetBillable() {
				row.billable += dur
			}
			row.entries++
		}
		if row.entries == 0 && row.total == 0 {
			continue
		}
		if onClock, timer := project.OnClock(); onClock {
			row.isOnClock = true
			row.clockAmount = timer.Duration()
		}
		rows = append(rows, row)
		overall += row.total
		overallBillable += row.billable
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].total > rows[j].total
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Report period: %s – %s\n\n", start.Format("2006-01-02"), end.Add(-time.Second).Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("%-28s %10s %10s %8s %s\n", "Project", "Total", "Billable", "Entries", "On clock"))
	sb.WriteString(strings.Repeat("-", 68))
	sb.WriteString("\n")

	for _, row := range rows {
		total := data.HmFromD(row.total).String()
		billable := data.HmFromD(row.billable).String()
		onClock := ""
		if row.isOnClock {
			onClock = fmt.Sprintf("running (%s)", data.HmFromD(row.clockAmount))
		}
		sb.WriteString(fmt.Sprintf("%-28s %10s %10s %8d %s\n", row.name, total, billable, row.entries, onClock))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Overall tracked: %s\n", data.HmFromD(overall)))
	sb.WriteString(fmt.Sprintf("Billable total: %s\n", data.HmFromD(overallBillable)))

	a.reportViewport.SetContent(sb.String())
	a.previousState = a.state
	a.state = stateReportView
}

// ProjectLogUI retains CLI log presentation capability within the TUI.
func (a *app) ProjectLogUI() {
	if a.project == nil {
		a.errorMessage = "No project selected"
		return
	}
	a.logViewport.SetContent(a.formatProjectLogs(a.project, a.logViewport.Width))
	if !a.logShowAll {
		a.logViewport.GotoTop()
	}
	if a.state != stateShowLogs {
		a.state = stateShowLogs
	}
}

// WebReplacementUI provides a dashboard in lieu of the old web view.
func (a *app) WebReplacementUI() {
	projects := data.DB.Projects()
	now := time.Now()
	weekStart := now.AddDate(0, 0, -6)
	type overview struct {
		name     string
		week     time.Duration
		month    time.Duration
		billable time.Duration
		onClock  bool
		clock    time.Duration
	}

	rows := make([]overview, 0, len(projects))
	var maxWeek time.Duration
	for _, project := range projects {
		var row overview
		row.name = project.GetName()
		for _, entry := range project.Entries() {
			ended, err := entry.EndedTime()
			if err != nil || ended == nil {
				continue
			}
			dur := time.Duration(entry.GetDuration())
			if ended.After(weekStart) {
				row.week += dur
			}
			if ended.Year() == now.Year() && ended.Month() == now.Month() {
				row.month += dur
			}
			if entry.GetBillable() {
				row.billable += dur
			}
		}
		if onClock, timer := project.OnClock(); onClock {
			row.onClock = true
			row.clock = timer.Duration()
		}
		if row.week > maxWeek {
			maxWeek = row.week
		}
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].week > rows[j].week })

	barWidth := 24
	if a.width > 0 {
		barWidth = a.width / 3
		if barWidth < 10 {
			barWidth = 10
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Weekly overview (since %s)\n\n", weekStart.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("%-20s %-8s %-8s %-8s %s\n", "Project", "7d", "Month", "Billable", "Activity"))
	sb.WriteString(strings.Repeat("-", 20+1+8+1+8+1+8+1+barWidth))
	sb.WriteString("\n")

	for _, row := range rows {
		bar := ""
		if maxWeek > 0 {
			ratio := float64(row.week) / float64(maxWeek)
			filled := int(ratio * float64(barWidth))
			if filled < 0 {
				filled = 0
			}
			if filled > barWidth {
				filled = barWidth
			}
			bar = strings.Repeat("█", filled) + strings.Repeat("·", barWidth-filled)
		}
		activity := ""
		if row.onClock {
			activity = fmt.Sprintf("running %s", data.HmFromD(row.clock))
		}
		line := fmt.Sprintf("%-20s %-8s %-8s %-8s %s %s\n",
			row.name,
			data.HmFromD(row.week),
			data.HmFromD(row.month),
			data.HmFromD(row.billable),
			activity,
			bar,
		)
		if a.project != nil && strings.EqualFold(row.name, a.project.GetName()) {
			sb.WriteString(selectedItemStyle.Render(line))
		} else {
			sb.WriteString(itemStyle.Render(line))
		}
	}

	a.dashboardViewport.SetContent(sb.String())
	a.previousState = a.state
	a.state = stateDashboard
}
