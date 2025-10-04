package tui

import (
	"fmt"
	"strings"
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
	titleStyle           = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle            = lipgloss.NewStyle().PaddingLeft(4)
	projectActionStyle   = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("250"))
	selectedItemStyle    = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle      = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle            = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1).Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	inputPromptStyle     = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("109")) // Style for input prompt
	errorStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).PaddingLeft(2) // Style for error messages
	logHeaderStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("37")).Bold(true)      // Style for log date headers
	logTotalStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Bold(true)      // Style for log totals
	logEntryStyle        = lipgloss.NewStyle()                                                  // Style for individual log entries
	logTitleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)      // Style for the main log title
	onClockStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))                 // Cool green for "on clock" status
	detailLabelStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("109")).Bold(true).PaddingLeft(2)
	detailValueStyle     = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("252"))
	detailSectionStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("111")).Bold(true).PaddingLeft(2)
	detailRowStyle       = detailValueStyle.Copy().PaddingLeft(2)
	detailHighlightStyle = detailRowStyle.Copy().Foreground(lipgloss.Color("213")).Bold(true)
	columnStyle          = lipgloss.NewStyle().Padding(0, 1)
	dividerStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// Define different states for the application
type state int

const (
	stateProjectList     state = iota // Showing the list of projects
	stateProjectMenu                  // Showing the menu for a selected project
	stateStoppingTimer                // Asking for a message before stopping timer
	stateManualEntry                  // Asking for time and message for manual entry
	stateShowLogs                     // Displaying project logs
	stateEntryList                    // Listing entries for a project
	stateConfirm                      // Generic confirmation prompt
	stateMoveEntryTarget              // Selecting target project for entry move
	stateRenameProject                // Renaming/moving a project
	stateReportView                   // Monthly report view
	stateDashboard                    // Overview/dashboard view
)

// Define focus states for manual entry
type manualFocus int

const (
	focusTime manualFocus = iota
	focusMessage
)

type confirmAction int

const (
	confirmNone confirmAction = iota
	confirmDeleteEntry
	confirmDeleteProject
)

type app struct {
	project           *data.Project
	projects          list.Model
	entries           list.Model
	choices           [][2]string
	state             state
	stopMessageInput  textinput.Model // Renamed for clarity
	manualTimeInput   textinput.Model // Input for manual entry time
	manualMsgInput    textinput.Model // Input for manual entry message
	manualEntryFocus  manualFocus     // Which input is focused in manual entry
	logViewport       viewport.Model  // Viewport for scrolling logs
	reportViewport    viewport.Model  // Viewport for report output
	dashboardViewport viewport.Model  // Viewport for dashboard output
	width             int             // store window width
	height            int             // store window height
	errorMessage      string          // To display temporary errors
	selectedEntry     *data.Entry
	confirmMessage    string
	confirmAction     confirmAction
	confirmEntry      *data.Entry
	confirmProject    *data.Project
	moveTargetProject *data.Project
	moveProjects      list.Model
	renameInput       textinput.Model
	reportMonth       time.Month
	reportYear        int
	logShowAll        bool
	previousState     state
}

func CreateApp() *app {
	var currentProject *data.Project
	projects := data.DB.Projects()
	if len(projects) > 0 {
		currentProject = projects[0]
	}
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
	l.SetShowHelp(false)
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

	renameTI := textinput.New()
	renameTI.Placeholder = "Enter new project name"
	renameTI.CharLimit = 120
	renameTI.Width = 50

	// Viewport for logs
	vp := viewport.New(defaultWidth, 20) // Initial size, will be updated
	vp.SetContent("Loading logs...")     // Placeholder content

	reportVP := viewport.New(defaultWidth, 20)
	reportVP.SetContent("Report will appear here")

	dashboardVP := viewport.New(defaultWidth, 20)
	dashboardVP.SetContent("Dashboard coming soon")

	initialState := stateProjectList
	if currentProject != nil {
		initialState = stateProjectMenu
	}

	a := &app{
		project:           currentProject,
		projects:          l,
		state:             initialState,
		stopMessageInput:  stopTI,
		manualTimeInput:   manualTimeTI,
		manualMsgInput:    manualMsgTI,
		manualEntryFocus:  focusTime,
		logViewport:       vp,
		reportViewport:    reportVP,
		dashboardViewport: dashboardVP,
		choices: [][2]string{
			{"s", "Start timer"},
			{"p", "End timer"},
			{"e", "Enter manually"},
			{"l", "Show logs"},
			{"v", "Entries"},
			{"D", "Delete project"},
			{"R", "Rename project"},
		},
		renameInput:   renameTI,
		reportMonth:   time.Now().Month(),
		reportYear:    time.Now().Year(),
		previousState: initialState,
	}

	if currentProject != nil {
		for idx, p := range projects {
			if p.GetName() == currentProject.GetName() {
				a.projects.Select(idx)
				break
			}
		}
	}

	a.updateProjectSelectionFromList()

	return a
}

func (a app) Init() tea.Cmd {
	return tea.ClearScreen
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
		listWidth, _ := a.projectColumnWidths(msg.Width)
		listHeight := msg.Height - 4
		if listHeight < 5 {
			listHeight = 5
		}
		a.projects.SetSize(listWidth, listHeight)
		// Update viewport size, leave some space for title and help
		headerHeight := lipgloss.Height(a.logTitleView())
		footerHeight := lipgloss.Height(a.logHelpView())
		a.logViewport.Width = msg.Width
		a.logViewport.Height = msg.Height - headerHeight - footerHeight
		a.reportViewport.Width = msg.Width
		a.reportViewport.Height = msg.Height - 4
		a.dashboardViewport.Width = msg.Width
		a.dashboardViewport.Height = msg.Height - 4
		if len(a.entries.Items()) > 0 {
			a.entries.SetSize(msg.Width, msg.Height-6)
		}
		if len(a.moveProjects.Items()) > 0 {
			a.moveProjects.SetSize(msg.Width, msg.Height-6)
		}
		a.renameInput.Width = msg.Width - 10
		// Adjust input widths dynamically if desired
		// a.stopMessageInput.Width = msg.Width - 10
		// a.manualMsgInput.Width = msg.Width - 30
		// Re-render logs if we are in that state, as width might affect wrapping
		if a.project != nil {
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
		case stateEntryList:
			m, c := a.handleKeypressEntryList(msg)
			return m, c
		case stateConfirm:
			m, c := a.handleKeypressConfirm(msg)
			return m, c
		case stateMoveEntryTarget:
			m, c := a.handleKeypressMoveEntryTarget(msg)
			return m, c
		case stateRenameProject:
			m, c := a.handleKeypressRenameProject(msg)
			return m, c
		case stateReportView:
			m, c := a.handleKeypressReportView(msg)
			return m, c
		case stateDashboard:
			m, c := a.handleKeypressDashboard(msg)
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
	case stateEntryList:
		a.entries, cmd = a.entries.Update(msg)
		cmds = append(cmds, cmd)
	case stateMoveEntryTarget:
		a.moveProjects, cmd = a.moveProjects.Update(msg)
		cmds = append(cmds, cmd)
	case stateRenameProject:
		a.renameInput, cmd = a.renameInput.Update(msg)
		cmds = append(cmds, cmd)
	case stateReportView:
		a.reportViewport, cmd = a.reportViewport.Update(msg)
		cmds = append(cmds, cmd)
	case stateDashboard:
		a.dashboardViewport, cmd = a.dashboardViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...) // Batch commands
}

func (a *app) refreshEntryList() {
	if a.project == nil {
		return
	}
	width := a.width
	if width == 0 {
		width = 80
	}
	const entryListHeight = 10
	listModel := buildEntryList(a.project, width, entryListHeight)
	a.entries = listModel
	a.selectedEntry = entryFromListItem(a.entries.SelectedItem())
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

func (a *app) exitConfirmation() {
	a.state = a.previousState
	a.confirmMessage = ""
	a.confirmAction = confirmNone
	a.confirmEntry = nil
	a.confirmProject = nil
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

func (a *app) adjustReportMonth(delta int) {
	month := int(a.reportMonth) + delta
	year := a.reportYear
	for month < 1 {
		month += 12
		year--
	}
	for month > 12 {
		month -= 12
		year++
	}
	now := time.Now()
	if year > now.Year() || (year == now.Year() && month > int(now.Month())) {
		year = now.Year()
		month = int(now.Month())
	}
	a.reportMonth = time.Month(month)
	a.reportYear = year
}

func truncateString(in string, max int) string {
	clean := strings.TrimSpace(strings.ReplaceAll(in, "\n", " "))
	if len(clean) <= max {
		return clean
	}
	if max <= 3 {
		return clean[:max]
	}
	return clean[:max-3] + "..."
}

func (a *app) entryDetailView(entry *data.Entry) string {
	lines := []string{titleStyle.MarginTop(1).Render("Entry details")}
	if entry == nil {
		lines = append(lines, detailValueStyle.Render("Select an entry to see details"))
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	started, _ := entry.StartedTime()
	ended, _ := entry.EndedTime()
	duration := data.HmFromD(time.Duration(entry.GetDuration()))

	startedStr := "—"
	if started != nil && !started.IsZero() {
		startedStr = started.Format("Jan 02 2006 15:04")
	}
	endedStr := "—"
	if ended != nil && !ended.IsZero() {
		endedStr = ended.Format("Jan 02 2006 15:04")
	}
	billableStr := "No"
	if entry.GetBillable() {
		billableStr = "Yes"
	}
	tags := "—"
	if len(entry.GetTags()) > 0 {
		tags = "#" + strings.Join(entry.GetTags(), " #")
	}

	lines = append(lines,
		detailLine("Started:", startedStr),
		detailLine("Ended:", endedStr),
		detailLine("Duration:", duration.String()),
		detailLine("Billable:", billableStr),
		detailLine("Tags:", tags),
	)

	desc := strings.TrimSpace(entry.GetContent())
	if desc == "" {
		desc = "(no description)"
	}

	lines = append(lines, "")
	lines = append(lines, detailSectionStyle.Render("Description"))
	for _, line := range strings.Split(desc, "\n") {
		lines = append(lines, detailValueStyle.Copy().PaddingLeft(3).Render(line))
	}

	lines = append(lines, "")
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func detailLine(label, value string) string {
	if value == "" {
		value = "—"
	}
	labelRendered := detailLabelStyle.Copy().Width(12).Render(label)
	return lipgloss.JoinHorizontal(lipgloss.Left, labelRendered, detailValueStyle.Render(value))
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
			choiceText = fmt.Sprintf("[%s] %s", choice[0], choice[1])
		} else {
			choiceText = fmt.Sprintf("    %s", choice[1])
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
	return lipgloss.JoinVertical(lipgloss.Left, top, "", a.projectFooterView())
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

func (a *app) handleKeypressEntryList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "esc":
		a.state = stateProjectMenu
		a.errorMessage = ""
		return a, nil
	case "m":
		entry := entryFromListItem(a.entries.SelectedItem())
		if entry == nil {
			return a, nil
		}
		a.selectedEntry = entry
		a.prepareMoveProjectList()
		if len(a.moveProjects.Items()) == 0 {
			a.errorMessage = "No other projects available to move this entry."
			return a, nil
		}
		a.previousState = stateEntryList
		a.state = stateMoveEntryTarget
		return a, nil
	case "d":
		entry := entryFromListItem(a.entries.SelectedItem())
		if entry != nil {
			a.selectedEntry = entry
			a.confirmEntry = entry
			a.confirmAction = confirmDeleteEntry
			a.confirmMessage = fmt.Sprintf("Delete entry '%s'?", truncateString(entry.GetContent(), 40))
			a.previousState = stateEntryList
			a.state = stateConfirm
		}
		return a, nil
	}

	var cmd tea.Cmd
	a.entries, cmd = a.entries.Update(msg)
	a.selectedEntry = entryFromListItem(a.entries.SelectedItem())
	return a, cmd
}

func (a *app) handleKeypressConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c":
		return a, tea.Quit
	case "esc", "n":
		a.exitConfirmation()
		return a, nil
	case "enter", "y":
		switch a.confirmAction {
		case confirmDeleteEntry:
			a.RemoveEntryUI()
		case confirmDeleteProject:
			a.RemoveProjectUI()
		}
		return a, nil
	}
	return a, nil
}

func (a *app) handleKeypressMoveEntryTarget(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "esc":
		a.state = a.previousState
		return a, nil
	case "enter":
		selected := a.moveProjects.SelectedItem()
		if projectItem, ok := selected.(item); ok {
			projects := lo.Filter(data.DB.Projects(), func(p *data.Project, _ int) bool {
				return *p.Name == string(projectItem)
			})
			if len(projects) > 0 {
				a.moveTargetProject = projects[0]
				a.MoveEntryUI()
			}
		}
		return a, nil
	}

	var cmd tea.Cmd
	a.moveProjects, cmd = a.moveProjects.Update(msg)
	return a, cmd
}

func (a *app) handleKeypressRenameProject(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "esc":
		a.renameInput.Blur()
		a.state = stateProjectMenu
		return a, nil
	case "enter":
		a.MoveProjectUI()
		return a, nil
	}

	var cmd tea.Cmd
	a.renameInput, cmd = a.renameInput.Update(msg)
	return a, cmd
}

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

func (a *app) handleKeypressDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keypress := msg.String(); keypress {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "esc":
		a.state = a.previousState
		return a, nil
	case "r":
		a.WebReplacementUI()
		return a, nil
	}

	var cmd tea.Cmd
	a.dashboardViewport, cmd = a.dashboardViewport.Update(msg)
	return a, cmd
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
		a.state = stateProjectMenu
		a.updateProjectSelectionFromList()
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

func (a app) View() string {
	var viewContent string

	switch a.state {
	case stateProjectList, stateProjectMenu:
		viewContent = a.projectSelectionView()

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

	case stateEntryList:
		if a.project == nil {
			viewContent = errorStyle.Render("No project selected")
		} else {
			projectName := *a.project.Name
			header := titleStyle.MarginTop(1).Render(fmt.Sprintf("Project: %s", projectName))
			entriesTitle := titleStyle.Render("Entries")
			help := helpStyle.Render("↑/↓: navigate | m: move entry | d: delete | esc: back | q: quit")
			entry := entryFromListItem(a.entries.SelectedItem())
			detail := a.entryDetailView(entry)
			viewContent = lipgloss.JoinVertical(lipgloss.Left,
				header,
				entriesTitle,
				a.entries.View(),
				detail,
				help,
			)
		}

	case stateConfirm:
		var lines []string
		lines = append(lines, titleStyle.MarginTop(1).Render("Please confirm"))
		lines = append(lines, "")
		lines = append(lines, itemStyle.Render(a.confirmMessage))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("y: confirm | n/esc: cancel | ctrl+c: quit"))
		viewContent = lipgloss.JoinVertical(lipgloss.Left, lines...)

	case stateMoveEntryTarget:
		viewContent = lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.MarginTop(1).Render("Move entry to which project?"),
			a.moveProjects.View(),
			helpStyle.Render("enter: move | esc: back | q: quit"),
		)

	case stateRenameProject:
		var lines []string
		lines = append(lines, titleStyle.MarginTop(1).Render("Rename project"))
		if a.project != nil {
			lines = append(lines, itemStyle.Render(fmt.Sprintf("Current name: %s", *a.project.Name)))
		}
		lines = append(lines, "")
		lines = append(lines, inputPromptStyle.Render(a.renameInput.View()))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("enter: save | esc: cancel | ctrl+c: quit"))
		viewContent = lipgloss.JoinVertical(lipgloss.Left, lines...)

	case stateReportView:
		title := titleStyle.MarginTop(1).Render(fmt.Sprintf("Monthly report: %s %d", a.reportMonth, a.reportYear))
		controls := helpStyle.Render("←/h: previous month | →/l: next month | r: reset | esc: back | q: quit")
		viewContent = lipgloss.JoinVertical(lipgloss.Left,
			title,
			a.reportViewport.View(),
			controls,
		)

	case stateDashboard:
		title := titleStyle.MarginTop(1).Render("Weekly overview")
		controls := helpStyle.Render("r: refresh | esc: back | q: quit")
		viewContent = lipgloss.JoinVertical(lipgloss.Left,
			title,
			a.dashboardViewport.View(),
			controls,
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
