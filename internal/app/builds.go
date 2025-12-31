package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elogrono/jenkins-tui/internal/browser"
	"github.com/elogrono/jenkins-tui/internal/jenkins"
	"github.com/elogrono/jenkins-tui/internal/jenkins/models"
	"github.com/elogrono/jenkins-tui/internal/ui/components"
	"github.com/elogrono/jenkins-tui/internal/ui/theme"
)

// BuildsMode represents what the builds tab is currently showing
type BuildsMode int

const (
	ModeJobList BuildsMode = iota
	ModeBuildList
	ModeBuildDetail
	ModeStageLogView
	ModeLogView
)

// BuildsModel handles the builds history tab
type BuildsModel struct {
	client *jenkins.Client
	width  int
	height int

	// Current mode
	mode BuildsMode

	// Data
	jobs          []models.Job
	selectedJob   int
	jobDetail     *models.JobDetail
	builds        []models.BuildRef
	selectedBuild int
	selectedStage int
	buildDetail   *models.Build
	pipelineRun   *models.PipelineRun
	logContent    string

	// UI components
	viewport    viewport.Model
	paginator   paginator.Model
	searchInput textinput.Model
	searching   bool
	filter      string

	// Pagination
	pageSize    int
	currentPage int
	totalBuilds int

	// Log viewing
	followLog      bool
	logSearchInput textinput.Model
	logSearching   bool
	logFilter      string

	// Scroll
	jobsScroll   int
	buildsScroll int

	// State
	loading    bool
	lastError  error
	lastUpdate time.Time
	spinner    spinner.Model
}

// NewBuildsModel creates a new builds model
func NewBuildsModel(client *jenkins.Client, width, height int) *BuildsModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.SpinnerStyle()

	search := textinput.New()
	search.Placeholder = "Type to search..."
	search.Width = 30
	search.Prompt = theme.IconSearch + " "

	logSearch := textinput.New()
	logSearch.Placeholder = "Search in log..."
	logSearch.Width = 30
	logSearch.Prompt = theme.IconSearch + " "

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 20

	vp := viewport.New(width, height-10)

	return &BuildsModel{
		client:         client,
		width:          width,
		height:         height,
		loading:        true,
		spinner:        s,
		searchInput:    search,
		logSearchInput: logSearch,
		paginator:      p,
		viewport:       vp,
		pageSize:       20,
		mode:           ModeJobList,
	}
}

// SetSize updates the dimensions
func (m *BuildsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width - 4
	m.viewport.Height = height - 12
}

// LoadData fetches builds data
func (m *BuildsModel) LoadData() tea.Cmd {
	m.loading = true
	return tea.Batch(
		m.fetchJobs(),
		m.spinner.Tick,
	)
}

// Update handles messages for the builds tab
func (m *BuildsModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case BuildsDataMsg:
		m.loading = false
		m.lastUpdate = time.Now()
		if msg.Jobs != nil {
			m.jobs = msg.Jobs
			// Sort by last build timestamp
			models.SortJobsByLastBuild(m.jobs)
		}
		if msg.JobDetail != nil {
			m.jobDetail = msg.JobDetail
			m.builds = msg.JobDetail.Builds
			// Sort builds by number (most recent first)
			models.SortBuildsByNumber(m.builds)
			m.totalBuilds = len(m.builds)
			m.paginator.SetTotalPages((m.totalBuilds + m.pageSize - 1) / m.pageSize)
		}
		if msg.BuildDetail != nil {
			m.buildDetail = msg.BuildDetail
		}
		if msg.PipelineRun != nil {
			m.pipelineRun = msg.PipelineRun
		}
		if msg.LogContent != "" {
			m.logContent = msg.LogContent
			m.viewport.SetContent(m.formatLogContent(m.logContent))
			if m.followLog {
				m.viewport.GotoBottom()
			}
		}
		if msg.Error != nil {
			m.lastError = msg.Error
		}
		return nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		// Handle search mode
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchInput.Blur()
				m.filter = ""
				m.searchInput.SetValue("")
				return nil
			case "enter":
				m.searching = false
				m.searchInput.Blur()
				m.filter = m.searchInput.Value()
				return nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return cmd
			}
		}

		// Handle log search mode
		if m.logSearching {
			switch msg.String() {
			case "esc":
				m.logSearching = false
				m.logSearchInput.Blur()
				// Keep the filter active, just exit search input mode
				// User can press Esc again to exit the log view
				return nil
			case "enter":
				m.logSearching = false
				m.logSearchInput.Blur()
				m.logFilter = m.logSearchInput.Value()
				m.highlightLogSearch()
				return nil
			default:
				var cmd tea.Cmd
				m.logSearchInput, cmd = m.logSearchInput.Update(msg)
				return cmd
			}
		}

		switch msg.String() {
		case "/":
			if m.mode == ModeLogView || m.mode == ModeStageLogView {
				m.logSearching = true
				m.logSearchInput.Focus()
				return textinput.Blink
			}
			m.searching = true
			m.searchInput.Focus()
			return textinput.Blink

		case "esc":
			switch m.mode {
			case ModeLogView, ModeStageLogView:
				m.mode = ModeBuildDetail
				m.logContent = ""
				m.logFilter = ""
			case ModeBuildDetail:
				m.mode = ModeBuildList
				m.buildDetail = nil
				m.pipelineRun = nil
			case ModeBuildList:
				m.mode = ModeJobList
				m.jobDetail = nil
				m.builds = nil
				m.selectedBuild = 0
				m.buildsScroll = 0
			}
			return nil

		case "enter":
			switch m.mode {
			case ModeJobList:
				filtered := m.getFilteredJobs()
				if len(filtered) > 0 && m.selectedJob < len(filtered) {
					m.mode = ModeBuildList
					return m.fetchJobDetail(filtered[m.selectedJob].Name)
				}
			case ModeBuildList:
				if len(m.builds) > 0 && m.selectedBuild < len(m.builds) {
					m.mode = ModeBuildDetail
					return m.fetchBuildDetail(m.jobDetail.Name, m.builds[m.selectedBuild].Number)
				}
			case ModeBuildDetail:
				if m.pipelineRun != nil && len(m.pipelineRun.Stages) > 0 {
					m.mode = ModeStageLogView
					stage := m.pipelineRun.Stages[m.selectedStage]
					return m.fetchStageLog(m.jobDetail.Name, m.buildDetail.Number, stage.ID)
				}
			}

		case "l":
			if (m.mode == ModeBuildList || m.mode == ModeBuildDetail) && m.jobDetail != nil {
				buildNum := 0
				if m.mode == ModeBuildList && len(m.builds) > 0 {
					buildNum = m.builds[m.selectedBuild].Number
				} else if m.mode == ModeBuildDetail && m.buildDetail != nil {
					buildNum = m.buildDetail.Number
				}
				if buildNum > 0 {
					m.mode = ModeLogView
					return m.fetchBuildLog(m.jobDetail.Name, buildNum)
				}
			}

		case "s":
			if m.mode == ModeLogView || m.mode == ModeStageLogView {
				m.followLog = !m.followLog
				if m.followLog {
					m.viewport.GotoBottom()
				}
			}

		case "r":
			switch m.mode {
			case ModeJobList:
				return m.LoadData()
			case ModeBuildList:
				if m.jobDetail != nil {
					return m.fetchJobDetail(m.jobDetail.Name)
				}
			case ModeLogView, ModeStageLogView:
				if m.jobDetail != nil && m.buildDetail != nil {
					if m.mode == ModeLogView {
						return m.fetchBuildLog(m.jobDetail.Name, m.buildDetail.Number)
					} else {
						stage := m.pipelineRun.Stages[m.selectedStage]
						return m.fetchStageLog(m.jobDetail.Name, m.buildDetail.Number, stage.ID)
					}
				}
			}

		case "o":
			// Open in browser
			var url string
			switch m.mode {
			case ModeJobList:
				filtered := m.getFilteredJobs()
				if len(filtered) > 0 && m.selectedJob < len(filtered) {
					url = filtered[m.selectedJob].URL
				}
			case ModeBuildList:
				if len(m.builds) > 0 && m.selectedBuild < len(m.builds) {
					url = m.builds[m.selectedBuild].URL
				}
			case ModeBuildDetail:
				if m.buildDetail != nil {
					url = m.buildDetail.URL
				}
			case ModeLogView, ModeStageLogView:
				if m.buildDetail != nil {
					url = m.buildDetail.URL + "console"
				}
			}
			if url != "" {
				_ = browser.Open(url)
			}
			return nil

		case "j", "down":
			m.navigateDown()

		case "k", "up":
			m.navigateUp()

		case "g":
			// Go to top
			m.selectedJob = 0
			m.selectedBuild = 0
			m.jobsScroll = 0
			m.buildsScroll = 0
			if m.mode == ModeLogView || m.mode == ModeStageLogView {
				m.viewport.GotoTop()
			}

		case "G":
			// Go to bottom
			switch m.mode {
			case ModeJobList:
				filtered := m.getFilteredJobs()
				if len(filtered) > 0 {
					m.selectedJob = len(filtered) - 1
				}
			case ModeBuildList:
				if len(m.builds) > 0 {
					m.selectedBuild = len(m.builds) - 1
				}
			case ModeLogView, ModeStageLogView:
				m.viewport.GotoBottom()
			}

		case "pgdown", "ctrl+f":
			if m.mode == ModeLogView || m.mode == ModeStageLogView {
				m.viewport.ViewDown()
			} else {
				m.pageDown()
			}

		case "pgup", "ctrl+b":
			if m.mode == ModeLogView || m.mode == ModeStageLogView {
				m.viewport.ViewUp()
			} else {
				m.pageUp()
			}
		}
	}

	return tea.Batch(cmds...)
}

// View renders the builds tab - FULL SCREEN
func (m *BuildsModel) View() string {
	if m.loading && len(m.jobs) == 0 {
		return m.viewLoading()
	}

	switch m.mode {
	case ModeJobList:
		return m.viewJobList()
	case ModeBuildList:
		return m.viewBuildList()
	case ModeBuildDetail:
		return m.viewBuildDetail()
	case ModeLogView, ModeStageLogView:
		return m.viewLog()
	}

	return ""
}

func (m *BuildsModel) viewLoading() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		m.spinner.View()+" Loading builds...",
		"",
		theme.MutedStyle.Render("Fetching data from Jenkins..."),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		theme.BoxStyle.Render(content))
}

func (m *BuildsModel) viewJobList() string {
	contentHeight := m.height - 8

	// Header
	breadcrumb := components.NewBreadcrumb("Builds").Render()
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		breadcrumb,
		strings.Repeat(" ", maxInt(0, m.width-lipgloss.Width(breadcrumb)-20)),
		theme.MutedStyle.Render(fmt.Sprintf("%d jobs", len(m.jobs))),
	)

	// Search bar
	var searchBar string
	if m.searching || m.filter != "" {
		searchBar = theme.SearchBarStyle.Width(m.width - 4).Render(m.searchInput.View())
	}

	// Column headers
	headerStyle := theme.TableHeaderStyle.Copy()
	colHeaders := fmt.Sprintf("  %-3s %-35s %-12s %-10s %-8s %-12s",
		"", "Job Name", "Last Build", "Result", "Health", "Updated")
	columnHeader := headerStyle.Width(m.width - 6).Render(colHeaders)

	// Jobs list
	filtered := m.getFilteredJobs()
	listHeight := contentHeight - 6
	if searchBar != "" {
		listHeight -= 3
	}

	var rows []string
	visibleStart := m.jobsScroll
	visibleEnd := minInt(visibleStart+listHeight, len(filtered))

	for i := visibleStart; i < visibleEnd; i++ {
		job := filtered[i]
		isSelected := i == m.selectedJob
		row := m.renderJobRow(job, isSelected, m.width-6)
		rows = append(rows, row)
	}

	list := strings.Join(rows, "\n")

	// Scroll info
	scrollInfo := ""
	if len(filtered) > listHeight {
		scrollInfo = theme.MutedStyle.Render(fmt.Sprintf(" [%d-%d of %d]", visibleStart+1, visibleEnd, len(filtered)))
	}

	// Shortcuts
	shortcuts := m.renderShortcuts()

	// Compose
	sections := []string{header}
	if searchBar != "" {
		sections = append(sections, searchBar)
	}
	sections = append(sections, columnHeader, list, scrollInfo, shortcuts)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-4).
		Padding(0, 1). // Minimal padding
		Render(content)
}

func (m *BuildsModel) viewBuildList() string {
	contentHeight := m.height - 8

	// Header
	jobName := ""
	if m.jobDetail != nil {
		jobName = m.jobDetail.Name
	}
	breadcrumb := components.NewBreadcrumb("Builds", jobName).Render()

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		breadcrumb,
		strings.Repeat(" ", maxInt(0, m.width-lipgloss.Width(breadcrumb)-25)),
		theme.MutedStyle.Render(fmt.Sprintf("%d builds", len(m.builds))),
	)

	// Extract stage names from first build that has stages
	stageNames := m.getStageNames()

	// Define column widths
	cols := m.calculateBuildColumns(stageNames)

	// Build column headers (potentially multi-line for stages)
	columnHeader := m.renderBuildTableHeader(cols, stageNames)

	// Builds list
	listHeight := contentHeight - 6

	var rows []string
	visibleStart := m.buildsScroll
	visibleEnd := minInt(visibleStart+listHeight, len(m.builds))

	for i := visibleStart; i < visibleEnd; i++ {
		build := m.builds[i]
		isSelected := i == m.selectedBuild
		row := m.renderBuildTableRow(build, isSelected, cols, stageNames)
		rows = append(rows, row)
	}

	list := strings.Join(rows, "\n")

	// Scroll info
	scrollInfo := ""
	if len(m.builds) > listHeight {
		scrollInfo = theme.MutedStyle.Render(fmt.Sprintf(" [%d-%d of %d]", visibleStart+1, visibleEnd, len(m.builds)))
	}

	// Shortcuts
	shortcuts := m.renderShortcuts()

	// Sections with NO spacing
	sections := []string{header, columnHeader, list, scrollInfo, shortcuts}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-4).
		Padding(0, 1).
		Render(content)
}

// BuildColumn defines a table column
type BuildColumn struct {
	Name  string
	Width int
}

// getStageNames extracts stage names from the first build that has stages
func (m *BuildsModel) getStageNames() []string {
	for _, build := range m.builds {
		if len(build.Stages) > 0 {
			names := make([]string, len(build.Stages))
			for i, stage := range build.Stages {
				names[i] = stage.Name
			}
			return names
		}
	}
	return nil
}

// calculateBuildColumns calculates column widths based on available space
func (m *BuildsModel) calculateBuildColumns(stageNames []string) []BuildColumn {
	availableWidth := m.width - 6 // Account for borders/padding

	// Fixed columns
	cols := []BuildColumn{
		{Name: "", Width: 3},         // Icon
		{Name: "Build", Width: 8},    // #number
		{Name: "Result", Width: 10},  // Status text
		{Name: "Date", Width: 17},    // YYYY-MM-DD HH:MM
		{Name: "Duration", Width: 9}, // Xm XXs
	}

	// Calculate used width
	usedWidth := 0
	for _, col := range cols {
		usedWidth += col.Width + 1 // +1 for separator
	}

	// Remaining width for stages
	remainingWidth := availableWidth - usedWidth
	numStages := len(stageNames)

	if numStages > 0 && remainingWidth > numStages*4 {
		// Calculate width per stage (minimum 4 chars)
		stageWidth := remainingWidth / numStages
		if stageWidth > 12 {
			stageWidth = 12 // Max stage column width
		}
		if stageWidth < 4 {
			stageWidth = 4
		}

		for _, name := range stageNames {
			cols = append(cols, BuildColumn{Name: name, Width: stageWidth})
		}
	}

	return cols
}

// renderBuildTableHeader renders the table header with proper alignment
func (m *BuildsModel) renderBuildTableHeader(cols []BuildColumn, stageNames []string) string {
	headerStyle := theme.TableHeaderStyle.Copy()

	var headerParts []string
	for i, col := range cols {
		cellContent := col.Name
		if len(cellContent) > col.Width {
			// Truncate with ellipsis
			cellContent = cellContent[:col.Width-1] + "."
		}

		// Center icons column, left-align others
		if i == 0 {
			// Icon column - centered
			padding := (col.Width - len(cellContent)) / 2
			cellContent = strings.Repeat(" ", padding) + cellContent + strings.Repeat(" ", col.Width-len(cellContent)-padding)
		} else {
			cellContent = fmt.Sprintf("%-*s", col.Width, cellContent)
		}

		headerParts = append(headerParts, cellContent)
	}

	headerLine := strings.Join(headerParts, " ")
	return headerStyle.Width(m.width - 6).Render(headerLine)
}

// renderBuildTableRow renders a single build row with proper column alignment
func (m *BuildsModel) renderBuildTableRow(build models.BuildRef, selected bool, cols []BuildColumn, stageNames []string) string {
	style := lipgloss.NewStyle().Width(m.width - 6)
	if selected {
		style = style.Background(theme.Primary).Foreground(theme.Background).Bold(true)
	}

	// Result
	result := build.Result
	if result == "" {
		result = "RUNNING"
	}

	// Build the row parts
	var rowParts []string

	for i, col := range cols {
		var cellContent string

		switch i {
		case 0: // Icon
			icon := theme.BuildResultIcon(result)
			// Center icon in cell
			cellContent = m.centerInWidth(icon, col.Width)

		case 1: // Build number
			cellContent = fmt.Sprintf("%-*s", col.Width, fmt.Sprintf("#%d", build.Number))

		case 2: // Result
			if selected {
				cellContent = fmt.Sprintf("%-*s", col.Width, result)
			} else {
				styled := theme.BuildResultStyle(result).Render(result)
				// Account for ANSI codes in width calculation
				padding := col.Width - len(result)
				if padding > 0 {
					cellContent = styled + strings.Repeat(" ", padding)
				} else {
					cellContent = styled
				}
			}

		case 3: // Date
			ts := time.UnixMilli(build.Timestamp)
			dateStr := ts.Format("2006-01-02 15:04")
			cellContent = fmt.Sprintf("%-*s", col.Width, dateStr)

		case 4: // Duration
			duration := time.Duration(build.Duration) * time.Millisecond
			durationStr := formatDuration(duration)
			cellContent = fmt.Sprintf("%-*s", col.Width, durationStr)

		default: // Stage columns
			stageIdx := i - 5
			if stageIdx < len(stageNames) {
				stageName := stageNames[stageIdx]
				icon := m.getStageIcon(build.Stages, stageName, selected)
				cellContent = m.centerInWidth(icon, col.Width)
			}
		}

		rowParts = append(rowParts, cellContent)
	}

	row := strings.Join(rowParts, " ")
	return style.Render(row)
}

// centerInWidth centers a string (possibly with ANSI codes) in the given width
func (m *BuildsModel) centerInWidth(s string, width int) string {
	// For icons, visual width is typically 1-2 chars
	visualWidth := 1 // Assume single char icon
	padding := (width - visualWidth) / 2
	if padding < 0 {
		padding = 0
	}
	rightPadding := width - visualWidth - padding
	if rightPadding < 0 {
		rightPadding = 0
	}
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", rightPadding)
}

// getStageIcon returns the icon for a specific stage in a build
func (m *BuildsModel) getStageIcon(stages []models.Stage, stageName string, selected bool) string {
	for _, s := range stages {
		if s.Name == stageName {
			status := s.Result
			if status == "" {
				status = s.Status
			}
			if selected {
				// Plain icon for selected row
				switch status {
				case "SUCCESS":
					return theme.IconSuccess
				case "FAILURE", "FAILED":
					return theme.IconFailure
				case "RUNNING", "IN_PROGRESS":
					return theme.IconRunning
				case "UNSTABLE":
					return theme.IconWarning
				case "ABORTED":
					return theme.IconAborted
				default:
					return theme.IconPending
				}
			}
			return theme.BuildResultIcon(status)
		}
	}
	// Stage not found in this build
	if selected {
		return "·"
	}
	return theme.MutedStyle.Render("·")
}

func (m *BuildsModel) viewBuildDetail() string {
	if m.buildDetail == nil {
		return m.viewLoading()
	}

	b := m.buildDetail

	// Header
	jobName := ""
	if m.jobDetail != nil {
		jobName = m.jobDetail.Name
	}
	breadcrumb := components.NewBreadcrumb("Builds", jobName, fmt.Sprintf("#%d", b.Number)).Render()

	// === Build Info Panel (Full width for simplicity and clarity) ===
	var infoRows []string

	// Status with badge
	status := b.StatusText()
	statusBadge := theme.StatusBadge(status)
	infoRows = append(infoRows, "  Status:      "+statusBadge)

	// Timestamp
	ts := time.UnixMilli(b.Timestamp)
	infoRows = append(infoRows, components.InfoRowWithIcon(theme.IconCalendar, "Started", ts.Format("2006-01-02 15:04:05")))

	// Duration
	duration := time.Duration(b.Duration) * time.Millisecond
	durationStr := formatDuration(duration)
	if b.Building {
		elapsed := time.Since(ts)
		durationStr = formatDuration(elapsed) + " (running)"
	}
	infoRows = append(infoRows, components.InfoRowWithIcon(theme.IconClock, "Duration", durationStr))

	// Progress (if running)
	if b.Building && b.EstimatedDuration > 0 {
		progress := b.GetProgress()
		progressBar := components.NewProgressBar(progress, 30).SetColor(theme.BuildRunning)
		infoRows = append(infoRows, "  Progress:    "+progressBar.Render())
	}

	// Trigger causes
	causes := b.GetCauses()
	if len(causes) > 0 {
		infoRows = append(infoRows, "")
		infoRows = append(infoRows, theme.SectionTitleStyle.Render(theme.IconUser+" Triggered by"))
		for _, cause := range causes {
			causeText := cause.ShortDescription
			if cause.UserName != "" {
				causeText = cause.UserName + " - " + cause.ShortDescription
			}
			infoRows = append(infoRows, "    "+theme.MutedStyle.Render(truncate(causeText, 80)))
		}
	}

	// Parameters
	params := b.GetParameters()
	if len(params) > 0 {
		infoRows = append(infoRows, "")
		infoRows = append(infoRows, theme.SectionTitleStyle.Render(theme.IconFilter+" Parameters"))
		for _, param := range params {
			valueStr := fmt.Sprintf("%v", param.Value)
			infoRows = append(infoRows, fmt.Sprintf("    %s = %s",
				theme.AccentStyle.Render(param.Name),
				theme.MutedStyle.Render(valueStr)))
		}
	}

	infoPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(m.width-6).
		Padding(1, 2).
		Render(strings.Join(infoRows, "\n"))

	// === Middle Section: Changes + Artifacts (Side by side) ===
	var changesRows []string
	changesRows = append(changesRows, theme.SectionTitleStyle.Render(theme.IconCommit+" Changes"))
	if len(b.ChangeSets) == 0 || (len(b.ChangeSets) > 0 && len(b.ChangeSets[0].Items) == 0) {
		changesRows = append(changesRows, theme.MutedStyle.Render("  No changes"))
	} else {
		for _, cs := range b.ChangeSets {
			for _, item := range cs.Items {
				msg := truncate(strings.TrimSpace(item.Msg), 40)
				author := truncate(item.Author.FullName, 15)
				commitID := ""
				if item.CommitID != "" {
					commitID = item.CommitID[:minInt(7, len(item.CommitID))]
				}
				changesRows = append(changesRows, fmt.Sprintf("  %s %s %s",
					theme.AccentStyle.Render(commitID),
					msg,
					theme.MutedStyle.Render("("+author+")")))
			}
		}
	}

	var artifactsRows []string
	artifactsRows = append(artifactsRows, theme.SectionTitleStyle.Render(theme.IconArtifact+" Artifacts"))
	if len(b.Artifacts) == 0 {
		artifactsRows = append(artifactsRows, theme.MutedStyle.Render("  No artifacts"))
	} else {
		for _, art := range b.Artifacts {
			artifactsRows = append(artifactsRows, fmt.Sprintf("  %s %s",
				theme.IconFile,
				truncate(art.FileName, 40)))
		}
	}

	changesPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(m.width/2-4).
		Padding(1, 2).
		Render(strings.Join(changesRows, "\n"))

	artifactsPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(m.width/2-4).
		Padding(1, 2).
		Render(strings.Join(artifactsRows, "\n"))

	middleContent := lipgloss.JoinHorizontal(lipgloss.Top, changesPanel, artifactsPanel)

	// === Bottom Section: Pipeline Stages (Full width) ===
	var stagesRows []string
	if m.pipelineRun != nil && len(m.pipelineRun.Stages) > 0 {
		stagesRows = append(stagesRows, theme.SectionTitleStyle.Render(theme.IconBuild+" Pipeline Stages"))
		for i, stage := range m.pipelineRun.Stages {
			status := components.StageStatusPending
			switch stage.Status {
			case "SUCCESS":
				status = components.StageStatusSuccess
			case "FAILED", "FAILURE":
				status = components.StageStatusFailure
			case "RUNNING", "IN_PROGRESS":
				status = components.StageStatusRunning
			case "UNSTABLE":
				status = components.StageStatusUnstable
			case "ABORTED":
				status = components.StageStatusAborted
			case "SKIPPED", "NOT_BUILT":
				status = components.StageStatusSkipped
			}

			prefix := "  "
			if i == m.selectedStage {
				prefix = theme.PrimaryStyle.Render("> ")
			}

			statusIcon := theme.MutedStyle.Render(string(status))
			switch status {
			case components.StageStatusSuccess:
				statusIcon = theme.SuccessStyle.Render(theme.IconSuccess)
			case components.StageStatusFailure:
				statusIcon = theme.ErrorStyle.Render(theme.IconFailure)
			case components.StageStatusRunning:
				statusIcon = theme.RunningStyle.Render(theme.IconRunning)
			}

			dur := formatDuration(time.Duration(stage.DurationMillis) * time.Millisecond)
			row := fmt.Sprintf("%s%s %-30s %s", prefix, statusIcon, truncate(stage.Name, 30), theme.MutedStyle.Render(dur))

			if i == m.selectedStage {
				row = theme.PrimaryStyle.Bold(true).Render(row)
			}
			stagesRows = append(stagesRows, row)
		}
	}

	stagesPanel := ""
	if len(stagesRows) > 0 {
		stagesPanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Width(m.width-6).
			Padding(1, 2).
			Render(strings.Join(stagesRows, "\n"))
	}

	// Combine all sections
	sections := []string{breadcrumb, infoPanel, middleContent}
	if stagesPanel != "" {
		sections = append(sections, stagesPanel)
	}

	// Shortcuts
	shortcuts := m.renderShortcuts()
	sections = append(sections, shortcuts)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-4).
		Padding(0, 1).
		Render(content)
}

func (m *BuildsModel) viewLog() string {
	// Header
	buildNum := 0
	jobName := ""
	if m.jobDetail != nil {
		jobName = m.jobDetail.Name
	}
	if m.buildDetail != nil {
		buildNum = m.buildDetail.Number
	}

	title := "Log"
	if m.mode == ModeStageLogView && m.pipelineRun != nil {
		title = "Stage: " + m.pipelineRun.Stages[m.selectedStage].Name
	}

	breadcrumb := components.NewBreadcrumb("Builds", jobName, fmt.Sprintf("#%d", buildNum), title).Render()

	// Status bar for log
	var statusItems []string
	statusItems = append(statusItems, theme.MutedStyle.Render(fmt.Sprintf("Lines: %d", strings.Count(m.logContent, "\n"))))
	statusItems = append(statusItems, theme.MutedStyle.Render(fmt.Sprintf("Size: %s", components.FormatBytes(int64(len(m.logContent))))))

	if m.followLog {
		statusItems = append(statusItems, theme.RunningStyle.Render("[FOLLOW ON]"))
	}

	if m.logFilter != "" {
		matches := strings.Count(strings.ToLower(m.logContent), strings.ToLower(m.logFilter))
		statusItems = append(statusItems, theme.AccentStyle.Render(fmt.Sprintf("Found: %d matches", matches)))
	}

	statusBar := strings.Join(statusItems, " │ ")

	// Search bar (if searching)
	var searchBar string
	if m.logSearching || m.logFilter != "" {
		searchBar = theme.SearchBarStyle.Width(m.width - 4).Render(m.logSearchInput.View())
	}

	// Update viewport size - ocupa TODO el ancho disponible
	m.viewport.Width = m.width - 2    // Quitar padding
	m.viewport.Height = m.height - 10 // Ajustar para breadcrumb, status, shortcuts
	if searchBar != "" {
		m.viewport.Height -= 2
	}

	// Log content SIN bordes para maximizar espacio
	logPanel := m.viewport.View()

	// Scroll position
	scrollInfo := theme.MutedStyle.Render(fmt.Sprintf(" %d%% ", int(m.viewport.ScrollPercent()*100)))

	// Shortcuts
	shortcuts := m.renderShortcuts()

	// Sections with NO spacing
	sections := []string{breadcrumb, statusBar}
	if searchBar != "" {
		sections = append(sections, searchBar)
	}
	sections = append(sections, logPanel, scrollInfo, shortcuts)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-4).
		Padding(0, 1). // Minimal padding
		Render(content)
}

func (m *BuildsModel) renderJobRow(job models.Job, selected bool, width int) string {
	style := lipgloss.NewStyle().Width(width)
	if selected {
		style = style.Background(theme.Primary).Foreground(theme.Background).Bold(true)
	}

	// Status icon
	statusIcon := theme.BuildStatusIcon(job.Color)

	// Job name
	name := truncate(job.Name, 33)

	// Last build
	lastBuild := "-"
	result := "-"
	timeAgo := "-"
	if job.LastBuild != nil {
		lastBuild = fmt.Sprintf("#%d", job.LastBuild.Number)
		result = job.LastBuild.Result
		if result == "" && job.IsRunning() {
			result = "RUNNING"
		}
		timeAgo = components.FormatTimeAgo(time.UnixMilli(job.LastBuild.Timestamp))
	}

	// Health
	health := "-"
	if len(job.HealthReport) > 0 {
		health = theme.HealthBar(job.HealthReport[0].Score)
	}

	// Format result with color (only if not selected)
	resultStr := result
	if !selected && result != "-" {
		resultStr = theme.BuildResultStyle(result).Render(result)
	}

	row := fmt.Sprintf("  %s %-35s %-12s %-10s %-8s %-12s",
		statusIcon, name, lastBuild, resultStr, health, timeAgo)

	return style.Render(row)
}

func (m *BuildsModel) renderShortcuts() string {
	bar := components.NewShortkeyBar(m.width)

	switch m.mode {
	case ModeJobList:
		bar.Add("/", "Search").
			Add("Enter", "View builds").
			Add("o", "Open URL").
			Add("r", "Refresh").
			Add("g/G", "Top/Bottom")
	case ModeBuildList:
		bar.Add("Enter", "Details").
			Add("l", "View log").
			Add("o", "Open URL").
			Add("Esc", "Back").
			Add("g/G", "Top/Bottom")
	case ModeBuildDetail:
		bar.Add("Enter", "Stage log").
			Add("l", "View full log").
			Add("o", "Open URL").
			Add("Esc", "Back")
	case ModeLogView, ModeStageLogView:
		bar.Add("/", "Search").
			Add("s", "Toggle follow").
			Add("o", "Open URL").
			Add("g/G", "Top/Bottom").
			Add("Esc", "Back")
	}

	// Add last update
	lastUpdate := ""
	if !m.lastUpdate.IsZero() {
		lastUpdate = theme.MutedStyle.Render(fmt.Sprintf(" │ Updated: %s", m.lastUpdate.Format("15:04:05")))
	}

	return bar.Render() + lastUpdate
}

func (m *BuildsModel) formatLogContent(content string) string {
	// Split lines and format with compact line numbers
	lines := strings.Split(content, "\n")
	var formatted []string

	// Calculate max line number width
	maxLineNum := len(lines)
	numWidth := len(fmt.Sprintf("%d", maxLineNum))

	for i, line := range lines {
		// Compact line number format
		lineNum := theme.MutedStyle.Render(fmt.Sprintf("%*d│ ", numWidth, i+1))

		// Colorize based on content
		styledLine := line
		lineLower := strings.ToLower(line)

		if strings.Contains(lineLower, "error") || strings.Contains(lineLower, "failed") ||
			strings.Contains(lineLower, "exception") {
			styledLine = theme.ErrorStyle.Render(line)
		} else if strings.Contains(lineLower, "warning") || strings.Contains(lineLower, "warn") {
			styledLine = theme.WarningStyle.Render(line)
		} else if strings.Contains(lineLower, "success") || strings.Contains(lineLower, "passed") {
			styledLine = theme.SuccessStyle.Render(line)
		} else if strings.HasPrefix(line, "[Pipeline]") || strings.HasPrefix(line, "[INFO]") {
			styledLine = theme.InfoStyle.Render(line)
		}

		// NO añadir más saltos de línea, solo concatenar
		formatted = append(formatted, lineNum+styledLine)
	}

	return strings.Join(formatted, "\n")
}

func (m *BuildsModel) highlightLogSearch() {
	if m.logFilter == "" {
		m.viewport.SetContent(m.formatLogContent(m.logContent))
		return
	}

	// First format, then highlight search term
	formatted := m.formatLogContent(m.logContent)

	// Simple case-insensitive highlight
	highlighted := strings.ReplaceAll(formatted, m.logFilter,
		theme.AccentStyle.Reverse(true).Render(m.logFilter))
	highlighted = strings.ReplaceAll(highlighted, strings.ToLower(m.logFilter),
		theme.AccentStyle.Reverse(true).Render(strings.ToLower(m.logFilter)))
	highlighted = strings.ReplaceAll(highlighted, strings.ToUpper(m.logFilter),
		theme.AccentStyle.Reverse(true).Render(strings.ToUpper(m.logFilter)))

	m.viewport.SetContent(highlighted)
}

func (m *BuildsModel) getFilteredJobs() []models.Job {
	if m.filter == "" {
		return m.jobs
	}

	var filtered []models.Job
	filterLower := strings.ToLower(m.filter)
	for _, job := range m.jobs {
		if strings.Contains(strings.ToLower(job.Name), filterLower) {
			filtered = append(filtered, job)
		}
	}
	return filtered
}

func (m *BuildsModel) navigateDown() {
	switch m.mode {
	case ModeJobList:
		filtered := m.getFilteredJobs()
		if m.selectedJob < len(filtered)-1 {
			m.selectedJob++
			// Update scroll to keep selection visible
			listHeight := m.height - 14
			if m.selectedJob >= m.jobsScroll+listHeight {
				m.jobsScroll = m.selectedJob - listHeight + 1
			}
		}
	case ModeBuildList:
		if m.selectedBuild < len(m.builds)-1 {
			m.selectedBuild++
			// Update scroll to keep selection visible
			listHeight := m.height - 14
			if m.selectedBuild >= m.buildsScroll+listHeight {
				m.buildsScroll = m.selectedBuild - listHeight + 1
			}
		}
	case ModeBuildDetail:
		if m.pipelineRun != nil && m.selectedStage < len(m.pipelineRun.Stages)-1 {
			m.selectedStage++
		}
	}
}

func (m *BuildsModel) navigateUp() {
	switch m.mode {
	case ModeJobList:
		if m.selectedJob > 0 {
			m.selectedJob--
			// Update scroll to keep selection visible
			if m.selectedJob < m.jobsScroll {
				m.jobsScroll = m.selectedJob
			}
		}
	case ModeBuildList:
		if m.selectedBuild > 0 {
			m.selectedBuild--
			// Update scroll to keep selection visible
			if m.selectedBuild < m.buildsScroll {
				m.buildsScroll = m.selectedBuild
			}
		}
	case ModeBuildDetail:
		if m.selectedStage > 0 {
			m.selectedStage--
		}
	}
}

func (m *BuildsModel) triggerBuild(jobName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := m.client.TriggerBuild(ctx, jobName)
		if err != nil {
			return BuildsDataMsg{Error: err}
		}
		// Return a message that triggers a refresh
		return BuildsDataMsg{Error: nil} // We could use a specific "Toast" or "Status" message later
	}
}

func (m *BuildsModel) pageDown() {
	pageSize := m.height / 3
	switch m.mode {
	case ModeJobList:
		filtered := m.getFilteredJobs()
		m.selectedJob = minInt(m.selectedJob+pageSize, len(filtered)-1)
		// Update scroll position
		listHeight := m.height - 14
		if m.selectedJob >= m.jobsScroll+listHeight {
			m.jobsScroll = m.selectedJob - listHeight + 1
		}
	case ModeBuildList:
		m.selectedBuild = minInt(m.selectedBuild+pageSize, len(m.builds)-1)
		// Update scroll position
		listHeight := m.height - 14
		if m.selectedBuild >= m.buildsScroll+listHeight {
			m.buildsScroll = m.selectedBuild - listHeight + 1
		}
	}
}

func (m *BuildsModel) pageUp() {
	pageSize := m.height / 3
	switch m.mode {
	case ModeJobList:
		m.selectedJob = maxInt(m.selectedJob-pageSize, 0)
		// Update scroll position
		if m.selectedJob < m.jobsScroll {
			m.jobsScroll = m.selectedJob
		}
	case ModeBuildList:
		m.selectedBuild = maxInt(m.selectedBuild-pageSize, 0)
		// Update scroll position
		if m.selectedBuild < m.buildsScroll {
			m.buildsScroll = m.selectedBuild
		}
	}
}

func (m *BuildsModel) fetchJobs() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		jobs, err := m.client.GetAllJobs(ctx)
		return BuildsDataMsg{Jobs: jobs, Error: err}
	}
}

func (m *BuildsModel) fetchJobDetail(jobName string) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		job, err := m.client.GetJob(ctx, jobName)
		return BuildsDataMsg{JobDetail: job, Error: err}
	}
}

func (m *BuildsModel) fetchBuildDetail(jobName string, buildNum int) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fetch build detail
		build, err := m.client.GetBuild(ctx, jobName, buildNum)
		if err != nil {
			return BuildsDataMsg{Error: err}
		}

		// Also try to fetch pipeline stages (may return nil if not a pipeline job)
		pipelineRun, _ := m.client.GetPipelineRun(ctx, jobName, buildNum)

		return BuildsDataMsg{BuildDetail: build, PipelineRun: pipelineRun, Error: nil}
	}
}

func (m *BuildsModel) fetchBuildLog(jobName string, buildNum int) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		log, err := m.client.GetBuildLog(ctx, jobName, buildNum, 500000) // 500KB max
		return BuildsDataMsg{LogContent: log, Error: err}
	}
}

func (m *BuildsModel) fetchStageLog(jobName string, buildNum int, stageID string) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		log, err := m.client.GetStageLog(ctx, jobName, buildNum, stageID)
		return BuildsDataMsg{LogContent: log, Error: err}
	}
}

// BuildsDataMsg carries builds data updates
type BuildsDataMsg struct {
	Jobs        []models.Job
	JobDetail   *models.JobDetail
	BuildDetail *models.Build
	PipelineRun *models.PipelineRun
	LogContent  string
	Error       error
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "< 1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func statusToColor(status string) string {
	switch status {
	case "SUCCESS":
		return "blue"
	case "FAILURE":
		return "red"
	case "UNSTABLE":
		return "yellow"
	case "ABORTED":
		return "aborted"
	case "RUNNING":
		return "blue_anime"
	default:
		return "notbuilt"
	}
}
