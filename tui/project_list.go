package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nexneo/samay/data"
	"github.com/samber/lo"
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

func (a *app) prepareMoveProjectList() {
	width := a.width
	if width == 0 {
		width = 80
	}
	height := a.height - 6
	if height < 10 {
		height = 10
	}
	items := make([]list.Item, 0)
	for _, project := range data.DB.Projects() {
		if a.project != nil && project.GetName() == a.project.GetName() {
			continue
		}
		items = append(items, item(*project.Name))
	}
	l := list.New(items, itemDelegate{}, width, height)
	l.Title = "Select target project"
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	a.moveProjects = l
}

func (a *app) refreshProjectList() {
	projects := data.DB.Projects()
	items := make([]list.Item, 0, len(projects))
	for _, p := range projects {
		items = append(items, item(*p.Name))
	}
	a.projects.SetItems(items)
	height := len(items)*2 + 5
	if height < 10 {
		height = 10
	}
	width := a.width
	if width == 0 {
		width = 40
	}
	a.projects.SetSize(width, height)
	a.updateProjectSelectionFromList()
}

func verticalDivider(height int) string {
	if height < 1 {
		height = 1
	}
	segments := make([]string, height)
	for i := range segments {
		segments[i] = "│"
	}
	return dividerStyle.Render(strings.Join(segments, "\n"))
}

func (a app) projectColumnWidths(totalWidth int) (int, int) {
	if totalWidth <= 0 {
		totalWidth = 80
	}
	dividerWidth := lipgloss.Width("│")
	available := totalWidth - dividerWidth
	if available < 2 {
		return 1, 1
	}
	left := available / 2
	right := available - left
	return left, right
}

func (a app) projectFooterView() string {
	baseControls := []string{"↑/↓: navigate", "r: monthly report (list)", "o: weekly overview", "q: quit"}
	return helpStyle.Render(strings.Join(baseControls, " | "))
}

func (a app) projectActionsView(width int) string {
	if width <= 0 {
		width = 40
	}
	if a.project == nil {
		lines := []string{titleStyle.Render("Project actions"), "", projectActionStyle.Render("Select a project to see available actions.")}
		return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	onclock, _ := a.project.OnClock()
	projectName := "project: " + *a.project.Name
	if onclock {
		projectName += onClockStyle.Render(" (on clock)")
	}
	lines := []string{
		titleStyle.Render(projectName),
		"",
	}
	for _, choice := range a.choices {
		if onclock && choice[0] == "s" {
			continue
		}
		if !onclock && choice[0] == "p" {
			continue
		}
		var choiceText string
		if choice[0] != "" {
			shortcut := projectShortcutStyle.Render(fmt.Sprintf("%s: ", choice[0]))
			label := projectLabelStyle.Render(choice[1])
			choiceText = lipgloss.JoinHorizontal(lipgloss.Left, shortcut, label)
		} else {
			label := projectLabelStyle.Render(choice[1])
			choiceText = lipgloss.JoinHorizontal(lipgloss.Left, projectShortcutSlot.Render(""), " ", label)
		}
		lines = append(lines, projectActionStyle.Render(choiceText))
	}
	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (a app) projectSelectionView() string {
	totalWidth := a.width
	if totalWidth <= 0 {
		totalWidth = 80
	}
	_, rightWidth := a.projectColumnWidths(totalWidth)
	leftContent := columnStyle.Render(a.projects.View())
	rightContent := columnStyle.Render(a.projectActionsView(rightWidth))
	height := lipgloss.Height(leftContent)
	if h := lipgloss.Height(rightContent); h > height {
		height = h
	}
	div := verticalDivider(height)
	top := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, div, rightContent)
	return lipgloss.JoinVertical(lipgloss.Left, "", top, a.projectFooterView())
}

func (a *app) updateProjectSelectionFromList() {
	items := a.projects.Items()
	if len(items) == 0 {
		a.project = nil
		a.state = stateProjectList
		return
	}

	selected := a.projects.SelectedItem()
	selectedItem, ok := selected.(item)
	if !ok {
		a.projects.Select(0)
		selected = a.projects.SelectedItem()
		selectedItem, ok = selected.(item)
		if !ok {
			a.project = nil
			a.state = stateProjectList
			return
		}
	}

	name := string(selectedItem)
	if project, found := lo.Find(data.DB.Projects(), func(p *data.Project) bool {
		return *p.Name == name
	}); found {
		a.project = project
		if a.state == stateProjectList {
			a.state = stateProjectMenu
		}
		return
	}

	firstItem, ok := items[0].(item)
	if !ok {
		a.project = nil
		a.state = stateProjectList
		return
	}
	if project, found := lo.Find(data.DB.Projects(), func(p *data.Project) bool {
		return *p.Name == string(firstItem)
	}); found {
		a.projects.Select(0)
		a.project = project
		if a.state == stateProjectList {
			a.state = stateProjectMenu
		}
		return
	}

	a.project = nil
	a.state = stateProjectList
}

// when the project list is active
func (a *app) handleKeypressProjectList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "r":
		a.ReportViewUI()
		return a, nil
	case "o":
		a.WebReplacementUI()
		return a, nil
	}

	// Default list navigation
	var cmd tea.Cmd
	a.projects, cmd = a.projects.Update(msg)
	a.updateProjectSelectionFromList()
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
	case "r":
		a.ReportViewUI()
		return a, nil
	case "o":
		a.WebReplacementUI()
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
			a.stopEntryFocus = focusStopMessage
			a.stopBillable = true
			a.stopMessageInput.Focus()
			a.stopMessageInput.SetValue("")
			return a, textinput.Blink
		}
		return a, nil
	case "e": // Enter Manually (Prepare)
		a.state = stateManualEntry
		a.manualEntryFocus = focusTime
		a.manualBillable = true
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
	case "v":
		a.refreshEntryList()
		a.selectedEntry = nil
		a.state = stateEntryList
		a.errorMessage = ""
		return a, nil
	case "D", "shift+d":
		a.confirmAction = confirmDeleteProject
		a.confirmProject = a.project
		a.confirmEntry = nil
		if a.project != nil {
			a.confirmMessage = fmt.Sprintf("Delete project '%s'? This removes all entries.", *a.project.Name)
		}
		a.previousState = stateProjectMenu
		a.state = stateConfirm
		return a, nil
	case "R", "shift+r":
		if a.project != nil {
			a.renameInput.SetValue(*a.project.Name)
		}
		a.state = stateRenameProject
		a.renameInput.Focus()
		return a, textinput.Blink
	}

	var cmd tea.Cmd
	a.projects, cmd = a.projects.Update(msg)
	a.updateProjectSelectionFromList()
	return a, cmd
}
