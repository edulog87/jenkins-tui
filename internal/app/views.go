package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elogrono/jenkins-tui/internal/browser"
	"github.com/elogrono/jenkins-tui/internal/jenkins"
	"github.com/elogrono/jenkins-tui/internal/jenkins/models"
	"github.com/elogrono/jenkins-tui/internal/ui/components"
	"github.com/elogrono/jenkins-tui/internal/ui/theme"
)

// ViewsMode represents the current mode in views tab
type ViewsMode int

const (
	ViewsModeList ViewsMode = iota
	ViewsModeJobs
	ViewsModeJobDetail
)

// ViewsModel handles the views tab
type ViewsModel struct {
	client *jenkins.Client
	width  int
	height int

	// Current mode
	mode ViewsMode

	// Data
	views        []models.View
	selectedView int
	jobs         []models.Job
	selectedJob  int
	jobDetail    *models.JobDetail

	// UI state
	searchInput textinput.Model
	searching   bool
	filter      string

	// Scroll state
	viewsScroll int
	jobsScroll  int

	// State
	loading    bool
	lastError  error
	lastUpdate time.Time
	spinner    spinner.Model
}

// NewViewsModel creates a new views model
func NewViewsModel(client *jenkins.Client, width, height int) *ViewsModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.SpinnerStyle()

	search := textinput.New()
	search.Placeholder = "Type to search..."
	search.Width = 30
	search.Prompt = theme.IconSearch + " "

	return &ViewsModel{
		client:      client,
		width:       width,
		height:      height,
		loading:     true,
		spinner:     s,
		searchInput: search,
		mode:        ViewsModeList,
	}
}

// SetSize updates the dimensions
func (m *ViewsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// LoadData fetches views data
func (m *ViewsModel) LoadData() tea.Cmd {
	m.loading = true
	return tea.Batch(
		m.fetchViews(),
		m.spinner.Tick,
	)
}

// Update handles messages for the views tab
func (m *ViewsModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ViewsDataMsg:
		m.loading = false
		m.lastUpdate = time.Now()
		if msg.Views != nil {
			m.views = msg.Views
		}
		if msg.Jobs != nil {
			m.jobs = msg.Jobs
			// Sort jobs by last build timestamp (most recent first)
			models.SortJobsByLastBuild(m.jobs)
		}
		if msg.JobDetail != nil {
			m.jobDetail = msg.JobDetail
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
				m.filter = m.searchInput.Value()
				return cmd
			}
		}

		switch msg.String() {
		case "/":
			m.searching = true
			m.searchInput.Focus()
			return textinput.Blink

		case "esc":
			switch m.mode {
			case ViewsModeJobDetail:
				m.mode = ViewsModeJobs
				m.jobDetail = nil
			case ViewsModeJobs:
				m.mode = ViewsModeList
				m.jobs = nil
				m.selectedJob = 0
				m.jobsScroll = 0
			}
			return nil

		case "enter":
			switch m.mode {
			case ViewsModeList:
				if len(m.getFilteredViews()) > 0 {
					m.mode = ViewsModeJobs
					return m.fetchViewJobs(m.getFilteredViews()[m.selectedView].Name)
				}
			case ViewsModeJobs:
				filteredJobs := m.getFilteredJobs()
				if len(filteredJobs) > 0 && m.selectedJob < len(filteredJobs) {
					m.mode = ViewsModeJobDetail
					return m.fetchJobDetail(filteredJobs[m.selectedJob].Name)
				}
			}

		case "o":
			// Open in browser
			var url string
			switch m.mode {
			case ViewsModeList:
				filtered := m.getFilteredViews()
				if len(filtered) > 0 && m.selectedView < len(filtered) {
					url = filtered[m.selectedView].URL
				}
			case ViewsModeJobs:
				filtered := m.getFilteredJobs()
				if len(filtered) > 0 && m.selectedJob < len(filtered) {
					url = filtered[m.selectedJob].URL
				}
			case ViewsModeJobDetail:
				if m.jobDetail != nil {
					url = m.jobDetail.URL
				}
			}
			if url != "" {
				_ = browser.Open(url)
			}
			return nil

		case "r":
			if m.mode == ViewsModeList {
				return m.LoadData()
			} else if m.mode == ViewsModeJobs {
				return m.fetchViewJobs(m.views[m.selectedView].Name)
			}

		case "j", "down":
			m.navigateDown()

		case "k", "up":
			m.navigateUp()

		case "g":
			// Go to top
			m.selectedView = 0
			m.selectedJob = 0
			m.viewsScroll = 0
			m.jobsScroll = 0

		case "G":
			// Go to bottom
			switch m.mode {
			case ViewsModeList:
				filtered := m.getFilteredViews()
				if len(filtered) > 0 {
					m.selectedView = len(filtered) - 1
				}
			case ViewsModeJobs:
				filtered := m.getFilteredJobs()
				if len(filtered) > 0 {
					m.selectedJob = len(filtered) - 1
				}
			}

		case "pgdown":
			m.pageDown()

		case "pgup":
			m.pageUp()

		case "backspace":
			if m.filter != "" {
				m.filter = ""
				m.searchInput.SetValue("")
			}
		}
	}

	return tea.Batch(cmds...)
}

// View renders the views tab - FULL SCREEN
func (m *ViewsModel) View() string {
	if m.loading && len(m.views) == 0 {
		return m.viewLoading()
	}

	switch m.mode {
	case ViewsModeList:
		return m.viewViewsList()
	case ViewsModeJobs:
		return m.viewJobsList()
	case ViewsModeJobDetail:
		return m.viewJobDetail()
	}

	return ""
}

func (m *ViewsModel) viewLoading() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		m.spinner.View()+" Loading views...",
		"",
		theme.MutedStyle.Render("Fetching data from Jenkins..."),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		theme.BoxStyle.Render(content))
}

func (m *ViewsModel) viewViewsList() string {
	contentHeight := m.height - 8

	// Header with breadcrumb
	breadcrumb := components.NewBreadcrumb("Views").Render()
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		breadcrumb,
		strings.Repeat(" ", maxInt(0, m.width-lipgloss.Width(breadcrumb)-20)),
		theme.MutedStyle.Render(fmt.Sprintf("%d views", len(m.views))),
	)

	// Search bar
	var searchBar string
	if m.searching || m.filter != "" {
		searchBar = theme.SearchBarStyle.Width(m.width - 4).Render(m.searchInput.View())
	}

	// Views list
	filtered := m.getFilteredViews()
	listHeight := contentHeight - 4
	if searchBar != "" {
		listHeight -= 3
	}

	var rows []string
	visibleStart := m.viewsScroll
	visibleEnd := minInt(visibleStart+listHeight, len(filtered))

	for i := visibleStart; i < visibleEnd; i++ {
		view := filtered[i]
		isSelected := i == m.selectedView

		row := m.renderViewRow(view, isSelected, m.width-6)
		rows = append(rows, row)
	}

	list := strings.Join(rows, "\n")

	// Scrollbar indicator
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
	sections = append(sections, list, scrollInfo, shortcuts)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-4).
		Padding(0, 1).
		Render(content)
}

func (m *ViewsModel) viewJobsList() string {
	contentHeight := m.height - 8

	// Header with breadcrumb
	viewName := ""
	if m.selectedView < len(m.views) {
		viewName = m.views[m.selectedView].Name
	}
	breadcrumb := components.NewBreadcrumb("Views", viewName).Render()

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
	colHeaders := fmt.Sprintf("  %-3s %-30s %-12s %-10s %-8s %-10s",
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

	// Scrollbar indicator
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
		Padding(0, 1).
		Render(content)
}

func (m *ViewsModel) viewJobDetail() string {
	if m.jobDetail == nil {
		return m.viewLoading()
	}

	job := m.jobDetail
	contentHeight := m.height - 8

	// Header
	breadcrumb := components.NewBreadcrumb("Views", m.views[m.selectedView].Name, job.Name).Render()

	// Job info panel
	var infoRows []string

	// Status row
	statusIcon := theme.BuildStatusIcon(job.Color)
	statusText := "Unknown"
	if job.LastBuild != nil {
		statusText = job.LastBuild.Result
		if statusText == "" && job.Color != "" && strings.Contains(job.Color, "_anime") {
			statusText = "RUNNING"
		}
	}
	infoRows = append(infoRows, components.InfoRowWithIcon(statusIcon, "Status", statusText))

	// Health
	healthScore := -1
	healthDesc := "Unknown"
	if len(job.HealthReport) > 0 {
		healthScore = job.HealthReport[0].Score
		healthDesc = job.HealthReport[0].Description
	}
	healthBar := ""
	if healthScore >= 0 {
		healthBar = theme.HealthBar(healthScore) + fmt.Sprintf(" %d%%", healthScore)
	}
	infoRows = append(infoRows, components.InfoRowWithIcon(theme.IconStar, "Health", healthBar))
	if healthDesc != "Unknown" {
		infoRows = append(infoRows, theme.MutedStyle.Render("         "+truncate(healthDesc, 50)))
	}

	// Last build
	if job.LastBuild != nil {
		buildNum := fmt.Sprintf("#%d", job.LastBuild.Number)
		timeAgo := components.FormatTimeAgo(time.UnixMilli(job.LastBuild.Timestamp))
		infoRows = append(infoRows, components.InfoRowWithIcon(theme.IconBuild, "Last Build", buildNum+" ("+timeAgo+")"))
	}

	// Description
	if job.Description != "" {
		desc := truncate(strings.TrimSpace(job.Description), 60)
		infoRows = append(infoRows, components.InfoRowWithIcon(theme.IconFile, "Description", desc))
	}

	// Buildable status
	buildableIcon := theme.IconSuccess
	buildableText := "Yes"
	if !job.Buildable {
		buildableIcon = theme.IconDisabled
		buildableText = "No (Disabled)"
	}
	infoRows = append(infoRows, components.InfoRowWithIcon(buildableIcon, "Buildable", buildableText))

	infoPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(m.width/2-4).
		Padding(1, 2).
		Render(strings.Join(infoRows, "\n"))

	// Recent builds panel
	buildsTitle := theme.SectionTitleStyle.Render(theme.IconBuild + " Recent Builds")
	var buildsRows []string

	maxBuilds := (contentHeight - 10) / 2
	if maxBuilds < 5 {
		maxBuilds = 5
	}

	// Sort builds by number (most recent first)
	builds := job.Builds
	models.SortBuildsByNumber(builds)

	for i, build := range builds {
		if i >= maxBuilds {
			buildsRows = append(buildsRows, theme.MutedStyle.Render(fmt.Sprintf("  ... and %d more builds", len(builds)-maxBuilds)))
			break
		}

		result := build.Result
		if result == "" {
			result = "RUNNING"
		}
		icon := theme.BuildResultIcon(result)
		timeAgo := components.FormatTimeAgo(time.UnixMilli(build.Timestamp))
		duration := formatDuration(time.Duration(build.Duration) * time.Millisecond)

		row := fmt.Sprintf("  %s #%-5d %-10s %-10s %s",
			icon,
			build.Number,
			theme.BuildResultStyle(result).Render(result),
			theme.MutedStyle.Render(duration),
			theme.MutedStyle.Render(timeAgo),
		)
		buildsRows = append(buildsRows, row)
	}

	buildsContent := strings.Join(buildsRows, "\n")
	buildsPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(m.width/2-4).
		Height(contentHeight-6).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, buildsTitle, "", buildsContent))

	// Layout
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, infoPanel, buildsPanel)

	// Shortcuts
	shortcuts := m.renderShortcuts()

	content := lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		"",
		mainContent,
		"",
		shortcuts,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-4).
		Padding(1, 2).
		Render(content)
}

func (m *ViewsModel) renderViewRow(view models.View, selected bool, width int) string {
	style := lipgloss.NewStyle().Width(width)
	if selected {
		style = style.Background(theme.Primary).Foreground(theme.Background).Bold(true)
	}

	icon := theme.IconFolder
	if selected {
		icon = theme.IconFolderOpen
	}

	// Show job count if available
	jobCount := view.JobCount()
	var row string
	if jobCount > 0 {
		if selected {
			row = fmt.Sprintf("  %s  %-35s (%d jobs)", icon, view.Name, jobCount)
		} else {
			row = fmt.Sprintf("  %s  %-35s %s", icon, view.Name, theme.MutedStyle.Render(fmt.Sprintf("(%d jobs)", jobCount)))
		}
	} else {
		row = fmt.Sprintf("  %s  %s", icon, view.Name)
	}
	return style.Render(row)
}

func (m *ViewsModel) renderJobRow(job models.Job, selected bool, width int) string {
	style := lipgloss.NewStyle().Width(width)
	if selected {
		style = style.Background(theme.Primary).Foreground(theme.Background).Bold(true)
	}

	// Status icon
	statusIcon := theme.BuildStatusIcon(job.Color)

	// Job name
	name := truncate(job.Name, 28)

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

	row := fmt.Sprintf("  %s %-30s %-12s %-10s %-8s %-10s",
		statusIcon, name, lastBuild, resultStr, health, timeAgo)

	return style.Render(row)
}

func (m *ViewsModel) renderShortcuts() string {
	bar := components.NewShortkeyBar(m.width)

	switch m.mode {
	case ViewsModeList:
		bar.Add("/", "Search").
			Add("Enter", "Open view").
			Add("r", "Refresh").
			Add("g/G", "Top/Bottom")
	case ViewsModeJobs:
		bar.Add("/", "Search").
			Add("Enter", "Job details").
			Add("Esc", "Back").
			Add("r", "Refresh").
			Add("g/G", "Top/Bottom")
	case ViewsModeJobDetail:
		bar.Add("Esc", "Back").
			Add("o", "Open in browser")
	}

	// Add last update
	lastUpdate := ""
	if !m.lastUpdate.IsZero() {
		lastUpdate = theme.MutedStyle.Render(fmt.Sprintf(" │ Updated: %s", m.lastUpdate.Format("15:04:05")))
	}

	return bar.Render() + lastUpdate
}

func (m *ViewsModel) getFilteredViews() []models.View {
	if m.filter == "" {
		return m.views
	}

	var filtered []models.View
	filterLower := strings.ToLower(m.filter)
	for _, view := range m.views {
		if strings.Contains(strings.ToLower(view.Name), filterLower) {
			filtered = append(filtered, view)
		}
	}
	return filtered
}

func (m *ViewsModel) getFilteredJobs() []models.Job {
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

func (m *ViewsModel) navigateDown() {
	switch m.mode {
	case ViewsModeList:
		filtered := m.getFilteredViews()
		if m.selectedView < len(filtered)-1 {
			m.selectedView++
			// Update scroll to keep selection visible
			listHeight := m.height - 12
			if m.selectedView >= m.viewsScroll+listHeight {
				m.viewsScroll = m.selectedView - listHeight + 1
			}
		}
	case ViewsModeJobs:
		filtered := m.getFilteredJobs()
		if m.selectedJob < len(filtered)-1 {
			m.selectedJob++
			// Update scroll to keep selection visible
			listHeight := m.height - 14
			if m.selectedJob >= m.jobsScroll+listHeight {
				m.jobsScroll = m.selectedJob - listHeight + 1
			}
		}
	}
}

func (m *ViewsModel) navigateUp() {
	switch m.mode {
	case ViewsModeList:
		if m.selectedView > 0 {
			m.selectedView--
			// Update scroll to keep selection visible
			if m.selectedView < m.viewsScroll {
				m.viewsScroll = m.selectedView
			}
		}
	case ViewsModeJobs:
		if m.selectedJob > 0 {
			m.selectedJob--
			// Update scroll to keep selection visible
			if m.selectedJob < m.jobsScroll {
				m.jobsScroll = m.selectedJob
			}
		}
	}
}

func (m *ViewsModel) pageDown() {
	pageSize := m.height / 3
	switch m.mode {
	case ViewsModeList:
		filtered := m.getFilteredViews()
		m.selectedView = minInt(m.selectedView+pageSize, len(filtered)-1)
		// Update scroll position
		listHeight := m.height - 12
		if m.selectedView >= m.viewsScroll+listHeight {
			m.viewsScroll = m.selectedView - listHeight + 1
		}
	case ViewsModeJobs:
		filtered := m.getFilteredJobs()
		m.selectedJob = minInt(m.selectedJob+pageSize, len(filtered)-1)
		// Update scroll position
		listHeight := m.height - 14
		if m.selectedJob >= m.jobsScroll+listHeight {
			m.jobsScroll = m.selectedJob - listHeight + 1
		}
	}
}

func (m *ViewsModel) pageUp() {
	pageSize := m.height / 3
	switch m.mode {
	case ViewsModeList:
		m.selectedView = maxInt(m.selectedView-pageSize, 0)
		// Update scroll position
		if m.selectedView < m.viewsScroll {
			m.viewsScroll = m.selectedView
		}
	case ViewsModeJobs:
		m.selectedJob = maxInt(m.selectedJob-pageSize, 0)
		// Update scroll position
		if m.selectedJob < m.jobsScroll {
			m.jobsScroll = m.selectedJob
		}
	}
}

func (m *ViewsModel) fetchViews() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		views, err := m.client.GetViews(ctx)
		return ViewsDataMsg{Views: views, Error: err}
	}
}

func (m *ViewsModel) fetchViewJobs(viewName string) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		jobs, err := m.client.GetViewJobs(ctx, viewName)
		return ViewsDataMsg{Jobs: jobs, Error: err}
	}
}

func (m *ViewsModel) fetchJobDetail(jobName string) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		job, err := m.client.GetJob(ctx, jobName)
		return ViewsDataMsg{JobDetail: job, Error: err}
	}
}

// ViewsDataMsg carries views data updates
type ViewsDataMsg struct {
	Views     []models.View
	Jobs      []models.Job
	JobDetail *models.JobDetail
	Error     error
}
