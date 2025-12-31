package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elogrono/jenkins-tui/internal/jenkins"
	"github.com/elogrono/jenkins-tui/internal/jenkins/models"
	"github.com/elogrono/jenkins-tui/internal/ui/components"
	"github.com/elogrono/jenkins-tui/internal/ui/theme"
)

// DashboardPanel represents which panel is selected
type DashboardPanel int

const (
	PanelRunning DashboardPanel = iota
	PanelNodes
	PanelQueue
	PanelRecent
)

// DashboardModel handles the dashboard tab
type DashboardModel struct {
	client *jenkins.Client
	width  int
	height int

	// Data
	rootInfo      *models.RootInfo
	nodes         []models.Node
	queue         *models.Queue
	runningBuilds []RunningBuildInfo
	recentBuilds  []RecentBuildInfo

	// Selection state
	selectedPanel     DashboardPanel
	selectedRunning   int
	selectedNode      int
	selectedQueueItem int
	selectedBuild     int

	// State
	loading    bool
	lastError  error
	lastUpdate time.Time
	spinner    spinner.Model
}

// RunningBuildInfo contains information about a running build
type RunningBuildInfo struct {
	JobName   string
	BuildNum  int
	NodeName  string
	StartTime time.Time
	Duration  time.Duration
	Progress  int
	URL       string
}

// RecentBuildInfo contains information about a recent build for dashboard display
type RecentBuildInfo struct {
	JobName   string
	BuildNum  int
	Result    string
	Color     string
	Timestamp time.Time
	Duration  time.Duration
	URL       string
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel(client *jenkins.Client, width, height int) *DashboardModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.SpinnerStyle()

	return &DashboardModel{
		client:        client,
		width:         width,
		height:        height,
		loading:       true,
		spinner:       s,
		selectedPanel: PanelRunning,
	}
}

// SetSize updates the dimensions
func (m *DashboardModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// LoadData fetches dashboard data
func (m *DashboardModel) LoadData() tea.Cmd {
	m.loading = true
	return tea.Batch(
		m.fetchRootInfo(),
		m.fetchNodes(),
		m.fetchQueue(),
		m.fetchJobs(),
		m.spinner.Tick,
	)
}

// Update handles messages for the dashboard
func (m *DashboardModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case DashboardDataMsg:
		m.loading = false
		m.lastUpdate = time.Now()
		if msg.RootInfo != nil {
			m.rootInfo = msg.RootInfo
		}
		if msg.Nodes != nil {
			m.nodes = msg.Nodes
			m.runningBuilds = m.extractRunningBuilds()
		}
		if msg.Queue != nil {
			m.queue = msg.Queue
		}
		if msg.Jobs != nil {
			m.processJobs(msg.Jobs)
		}
		if msg.Error != nil {
			m.lastError = msg.Error
		}
		return nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return cmd
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			return m.LoadData()
		case "j", "down":
			m.navigateDown()
		case "k", "up":
			m.navigateUp()
		case "tab":
			m.nextPanel()
		case "shift+tab":
			m.prevPanel()
		case "g":
			m.goToTop()
		case "G":
			m.goToBottom()
		}
	}
	return nil
}

// Panel navigation
func (m *DashboardModel) nextPanel() {
	m.selectedPanel = DashboardPanel((int(m.selectedPanel) + 1) % 4)
}

func (m *DashboardModel) prevPanel() {
	m.selectedPanel = DashboardPanel((int(m.selectedPanel) + 3) % 4)
}

func (m *DashboardModel) navigateDown() {
	switch m.selectedPanel {
	case PanelRunning:
		if m.selectedRunning < len(m.runningBuilds)-1 {
			m.selectedRunning++
		}
	case PanelNodes:
		if m.selectedNode < len(m.nodes)-1 {
			m.selectedNode++
		}
	case PanelQueue:
		if m.queue != nil && m.selectedQueueItem < len(m.queue.Items)-1 {
			m.selectedQueueItem++
		}
	case PanelRecent:
		if m.selectedBuild < len(m.recentBuilds)-1 {
			m.selectedBuild++
		}
	}
}

func (m *DashboardModel) navigateUp() {
	switch m.selectedPanel {
	case PanelRunning:
		if m.selectedRunning > 0 {
			m.selectedRunning--
		}
	case PanelNodes:
		if m.selectedNode > 0 {
			m.selectedNode--
		}
	case PanelQueue:
		if m.selectedQueueItem > 0 {
			m.selectedQueueItem--
		}
	case PanelRecent:
		if m.selectedBuild > 0 {
			m.selectedBuild--
		}
	}
}

func (m *DashboardModel) goToTop() {
	switch m.selectedPanel {
	case PanelRunning:
		m.selectedRunning = 0
	case PanelNodes:
		m.selectedNode = 0
	case PanelQueue:
		m.selectedQueueItem = 0
	case PanelRecent:
		m.selectedBuild = 0
	}
}

func (m *DashboardModel) goToBottom() {
	switch m.selectedPanel {
	case PanelRunning:
		if len(m.runningBuilds) > 0 {
			m.selectedRunning = len(m.runningBuilds) - 1
		}
	case PanelNodes:
		if len(m.nodes) > 0 {
			m.selectedNode = len(m.nodes) - 1
		}
	case PanelQueue:
		if m.queue != nil && len(m.queue.Items) > 0 {
			m.selectedQueueItem = len(m.queue.Items) - 1
		}
	case PanelRecent:
		if len(m.recentBuilds) > 0 {
			m.selectedBuild = len(m.recentBuilds) - 1
		}
	}
}

// View renders the dashboard - FULL SCREEN
func (m *DashboardModel) View() string {
	if m.loading && m.rootInfo == nil {
		return m.viewLoading()
	}

	// Calculate heights for layout components
	kpiHeight := 5
	shortcutHeight := 1
	availableHeight := m.height - kpiHeight - shortcutHeight - 2 // -2 for some breathing room

	if availableHeight < 10 {
		availableHeight = 10 // Safety minimum
	}

	// 1. KPI Row (Top)
	kpiRow := m.renderKPIs()

	// 2. Middle Section (Two columns)
	leftColWidth := m.width / 2
	rightColWidth := m.width - leftColWidth

	// Each column has two panels stacked vertically.
	// We want to give Running and Queue priority (e.g., 7 lines each including borders)
	topPanelsHeight := 7
	bottomPanelsHeight := availableHeight - topPanelsHeight

	if bottomPanelsHeight < 5 {
		// If the screen is very small, split 50/50
		topPanelsHeight = availableHeight / 2
		bottomPanelsHeight = availableHeight - topPanelsHeight
	}

	// Left column: Running (Top) | Nodes (Bottom)
	runningPanel := m.renderRunningBuildsPanel(leftColWidth, topPanelsHeight)
	nodesPanel := m.renderNodesPanel(leftColWidth, bottomPanelsHeight)
	leftCol := lipgloss.JoinVertical(lipgloss.Left, runningPanel, nodesPanel)

	// Right column: Queue (Top) | Recent Builds (Bottom)
	queuePanel := m.renderQueuePanel(rightColWidth, topPanelsHeight)
	recentPanel := m.renderRecentBuildsPanel(rightColWidth, bottomPanelsHeight)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, queuePanel, recentPanel)

	// Horizontal join of columns
	middleSection := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)

	// 3. Shortcuts (Bottom)
	shortcuts := m.renderShortcuts()

	// Final vertical join
	return lipgloss.JoinVertical(lipgloss.Left,
		kpiRow,
		middleSection,
		shortcuts,
	)
}

func (m *DashboardModel) viewLoading() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		m.spinner.View()+" Loading dashboard...",
		"",
		theme.MutedStyle.Render("Fetching data from Jenkins..."),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		theme.BoxStyle.Render(content))
}

func (m *DashboardModel) renderKPIs() string {
	// Calculate KPIs
	runningCount := len(m.runningBuilds)
	queueCount := 0
	queueBlocked := 0
	if m.queue != nil {
		queueCount = len(m.queue.Items)
		for _, item := range m.queue.Items {
			if item.Blocked {
				queueBlocked++
			}
		}
	}

	nodesOnline := 0
	nodesOffline := 0
	totalExecutors := 0
	busyExecutors := 0
	for _, node := range m.nodes {
		if node.Offline {
			nodesOffline++
		} else {
			nodesOnline++
			totalExecutors += node.NumExecutors
			busyExecutors += node.BusyExecutors()
		}
	}

	// Count failed from recent builds
	failedCount := 0
	for _, b := range m.recentBuilds {
		if b.Result == "FAILURE" || b.Color == "red" || b.Color == "red_anime" {
			failedCount++
		}
	}

	// Calculate card width based on screen
	cardWidth := (m.width - 10) / 5
	if cardWidth < 16 {
		cardWidth = 16
	}
	if cardWidth > 24 {
		cardWidth = 24
	}

	// Create KPI cards
	cards := []string{
		components.NewKPICard(
			"Running",
			fmt.Sprintf("%d", runningCount),
			theme.IconBuild,
			theme.BuildRunning,
		).SetWidth(cardWidth).Render(),

		components.NewKPICard(
			"Queue",
			fmt.Sprintf("%d", queueCount),
			theme.IconQueue,
			func() lipgloss.Color {
				if queueBlocked > 0 {
					return theme.Warning
				}
				return theme.Info
			}(),
		).SetWidth(cardWidth).Render(),

		components.NewKPICard(
			"Nodes",
			fmt.Sprintf("%d/%d", nodesOnline, nodesOnline+nodesOffline),
			theme.IconServer,
			func() lipgloss.Color {
				if nodesOffline > 0 {
					return theme.Warning
				}
				return theme.Success
			}(),
		).SetWidth(cardWidth).Render(),

		components.NewKPICard(
			"Executors",
			fmt.Sprintf("%d/%d", busyExecutors, totalExecutors),
			theme.IconRunning,
			theme.Primary,
		).SetWidth(cardWidth).Render(),

		components.NewKPICard(
			"Failed",
			fmt.Sprintf("%d", failedCount),
			theme.IconFailure,
			func() lipgloss.Color {
				if failedCount > 0 {
					return theme.Error
				}
				return theme.Success
			}(),
		).SetWidth(cardWidth).Render(),
	}

	return lipgloss.NewStyle().MarginBottom(0).Render(lipgloss.JoinHorizontal(lipgloss.Top, cards...))
}

func (m *DashboardModel) renderRunningBuildsPanel(width, height int) string {
	titleStyle := theme.SectionTitleStyle.Copy().Width(width - 2)
	title := titleStyle.Render(theme.IconBuild + " Running")

	var content string
	if len(m.runningBuilds) == 0 {
		content = theme.MutedStyle.Render("  No builds running")
	} else {
		var rows []string
		maxRows := height - 4 // -1 for header
		if maxRows < 1 {
			maxRows = 1
		}

		// Table Header
		header := theme.TableHeaderStyle.Copy().Width(width - 2).Render(
			fmt.Sprintf("  %-6s %-60s %-12s %-15s", "Build", "Job Name", "Progress", "Node"),
		)
		rows = append(rows, header)

		for i, build := range m.runningBuilds {
			if i >= maxRows-1 {
				remaining := len(m.runningBuilds) - (maxRows - 1)
				if remaining > 0 {
					rows = append(rows, theme.MutedStyle.Render(fmt.Sprintf("  ... and %d more", remaining)))
				}
				break
			}

			isSelected := m.selectedPanel == PanelRunning && i == m.selectedRunning

			// Progress bar
			progressBar := components.NewProgressBar(build.Progress, 12).SetColor(theme.BuildRunning)

			// Build info with node name
			jobName := build.JobName
			nodeName := build.NodeName

			var row string
			if isSelected {
				row = lipgloss.NewStyle().
					Background(theme.Primary).
					Foreground(theme.Background).
					Bold(true).
					Width(width - 2).
					Render(fmt.Sprintf(" %s#%-5d %-60s %-12s %-15s",
						theme.IconRunning,
						build.BuildNum,
						truncate(jobName, 60),
						progressBar.Render(),
						truncate(nodeName, 15),
					))
			} else {
				row = fmt.Sprintf("  %s#%-5d %-60s %-12s %s",
					theme.RunningStyle.Render(theme.IconRunning),
					build.BuildNum,
					theme.BaseStyle.Render(truncate(jobName, 60)),
					progressBar.Render(),
					theme.MutedStyle.Render(truncate(nodeName, 15)),
				)
			}
			rows = append(rows, row)
		}
		content = strings.Join(rows, "\n")
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(width).
		Height(height).
		Padding(0, 0) // Removed padding to keep it compact

	if m.selectedPanel == PanelRunning {
		panelStyle = panelStyle.BorderForeground(theme.Primary)
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m *DashboardModel) renderNodesPanel(width, height int) string {
	titleStyle := theme.SectionTitleStyle.Copy().Width(width - 2)
	title := titleStyle.Render(theme.IconServer + " Nodes")

	var content string
	if len(m.nodes) == 0 {
		content = theme.MutedStyle.Render("  No nodes found")
	} else {
		var rows []string
		maxRows := height - 4
		if maxRows < 1 {
			maxRows = 1
		}

		// Table Header
		header := theme.TableHeaderStyle.Copy().Width(width - 2).Render(
			fmt.Sprintf("  %-20s %-8s %-15s", "Name", "Exec", "Running"),
		)
		rows = append(rows, header)

		for i, node := range m.nodes {
			if i >= maxRows-1 { // -1 for header
				remaining := len(m.nodes) - (maxRows - 1)
				if remaining > 0 {
					rows = append(rows, theme.MutedStyle.Render(fmt.Sprintf("  ... and %d more", remaining)))
				}
				break
			}

			isSelected := m.selectedPanel == PanelNodes && i == m.selectedNode

			var statusIcon string
			var statusStyle lipgloss.Style
			if node.Offline {
				statusIcon = theme.IconFailure
				statusStyle = theme.ErrorStyle
			} else {
				statusIcon = theme.IconSuccess
				statusStyle = theme.SuccessStyle
			}

			execInfo := fmt.Sprintf("%d/%d", node.BusyExecutors(), node.NumExecutors)
			name := node.DisplayName

			// Show what's running on this node
			runningJob := ""
			for _, exec := range node.Executors {
				if exec.CurrentExecutable.URL != "" {
					jobName := exec.CurrentExecutable.FullDisplayName
					if jobName == "" {
						jobName = exec.CurrentExecutable.DisplayName
					}
					if jobName != "" {
						runningJob = jobName
						break
					}
				}
			}

			var row string
			if isSelected {
				row = lipgloss.NewStyle().
					Background(theme.Primary).
					Foreground(theme.Background).
					Bold(true).
					Width(width - 2).
					Render(fmt.Sprintf(" %s %-20s %-8s %-15s", statusIcon, truncate(name, 20), execInfo, truncate(runningJob, 15)))
			} else {
				runningStr := truncate(runningJob, 15)
				row = fmt.Sprintf("  %s %-20s %-8s %s",
					statusStyle.Render(statusIcon),
					truncate(name, 20),
					theme.MutedStyle.Render(execInfo),
					theme.AccentStyle.Render(runningStr),
				)
			}
			rows = append(rows, row)
		}
		content = strings.Join(rows, "\n")
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(width).
		Height(height).
		Padding(0, 0)

	if m.selectedPanel == PanelNodes {
		panelStyle = panelStyle.BorderForeground(theme.Primary)
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m *DashboardModel) renderQueuePanel(width, height int) string {
	titleStyle := theme.SectionTitleStyle.Copy().Width(width - 2)
	title := titleStyle.Render(theme.IconQueue + " Queue")

	var content string
	if m.queue == nil || len(m.queue.Items) == 0 {
		content = theme.MutedStyle.Render("  Empty")
	} else {
		var rows []string
		maxRows := height - 4 // -1 for header
		if maxRows < 1 {
			maxRows = 1
		}

		// Table Header
		header := theme.TableHeaderStyle.Copy().Width(width - 2).Render(
			fmt.Sprintf("  %-3s %-60s %s", "!", "Job Name", "Wait Time"),
		)
		rows = append(rows, header)

		for i, item := range m.queue.Items {
			if i >= maxRows-1 {
				remaining := len(m.queue.Items) - (maxRows - 1)
				if remaining > 0 {
					rows = append(rows, theme.MutedStyle.Render(fmt.Sprintf("  ... and %d more", remaining)))
				}
				break
			}

			isSelected := m.selectedPanel == PanelQueue && i == m.selectedQueueItem

			var statusIcon string
			var statusStyle lipgloss.Style
			if item.Stuck {
				statusIcon = theme.IconFailure
				statusStyle = theme.ErrorStyle
			} else if item.Blocked {
				statusIcon = theme.IconWarning
				statusStyle = theme.WarningStyle
			} else {
				statusIcon = theme.IconPending
				statusStyle = theme.MutedStyle
			}

			waitTime := components.FormatTimeAgo(time.UnixMilli(item.InQueueSince))
			name := item.Task.Name

			var row string
			if isSelected {
				row = lipgloss.NewStyle().
					Background(theme.Primary).
					Foreground(theme.Background).
					Bold(true).
					Width(width - 2).
					Render(fmt.Sprintf(" %s %-60s %s", statusIcon, truncate(name, 60), waitTime))
			} else {
				row = fmt.Sprintf("  %s %-60s %s",
					statusStyle.Render(statusIcon),
					truncate(name, 60),
					theme.MutedStyle.Render(waitTime),
				)
			}
			rows = append(rows, row)
		}
		content = strings.Join(rows, "\n")
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(width).
		Height(height).
		Padding(0, 0)

	if m.selectedPanel == PanelQueue {
		panelStyle = panelStyle.BorderForeground(theme.Primary)
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m *DashboardModel) renderRecentBuildsPanel(width, height int) string {
	titleStyle := theme.SectionTitleStyle.Copy().Width(width - 2)
	title := titleStyle.Render(theme.IconBuild + " Recent")

	var content string
	if len(m.recentBuilds) == 0 {
		content = theme.MutedStyle.Render("  No builds")
	} else {
		var rows []string
		maxRows := height - 4
		if maxRows > 15 {
			maxRows = 15
		}
		if maxRows < 1 {
			maxRows = 1
		}

		// Table Header
		header := theme.TableHeaderStyle.Copy().Width(width - 2).Render(
			fmt.Sprintf("  %-6s %-60s %-10s", "Build", "Job Name", "Time"),
		)
		rows = append(rows, header)

		for i, build := range m.recentBuilds {
			if i >= maxRows-1 { // -1 for header
				remaining := len(m.recentBuilds) - (maxRows - 1)
				if remaining > 0 {
					rows = append(rows, theme.MutedStyle.Render(fmt.Sprintf("  ... and %d more", remaining)))
				}
				break
			}

			isSelected := m.selectedPanel == PanelRecent && i == m.selectedBuild

			icon := theme.BuildStatusIcon(build.Color)
			name := build.JobName
			timeAgo := components.FormatTimeAgo(build.Timestamp)

			var row string
			if isSelected {
				row = lipgloss.NewStyle().
					Background(theme.Primary).
					Foreground(theme.Background).
					Bold(true).
					Width(width - 2).
					Render(fmt.Sprintf(" %s #%-5d %-60s %s", icon, build.BuildNum, truncate(name, 60), timeAgo))
			} else {
				resultStyle := theme.BuildResultStyle(build.Result)
				row = fmt.Sprintf("  %s #%-5d %-60s %s",
					icon,
					build.BuildNum,
					resultStyle.Render(truncate(name, 60)),
					theme.MutedStyle.Render(timeAgo),
				)
			}
			rows = append(rows, row)
		}
		content = strings.Join(rows, "\n")
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(width).
		Height(height).
		Padding(0, 0)

	if m.selectedPanel == PanelRecent {
		panelStyle = panelStyle.BorderForeground(theme.Primary)
	}

	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m *DashboardModel) renderShortcuts() string {
	bar := components.NewShortkeyBar(m.width).
		Add("r", "Refresh").
		Add("Tab", "Switch panel").
		Add("j/k", "Navigate").
		Add("g/G", "Top/Bottom").
		Add("?", "Help")

	// Add last update time
	lastUpdate := ""
	if !m.lastUpdate.IsZero() {
		lastUpdate = theme.MutedStyle.Render(fmt.Sprintf(" │ Updated: %s", m.lastUpdate.Format("15:04:05")))
	}

	return bar.Render() + lastUpdate
}

func (m *DashboardModel) extractRunningBuilds() []RunningBuildInfo {
	var running []RunningBuildInfo
	for _, node := range m.nodes {
		for _, exec := range node.Executors {
			if exec.CurrentExecutable.URL != "" {
				// Parse job name from URL
				jobName := exec.CurrentExecutable.FullDisplayName
				if jobName == "" {
					jobName = exec.CurrentExecutable.DisplayName
				}
				if jobName == "" {
					jobName = "Unknown"
				}

				// Calculate progress
				progress := exec.Progress
				if progress == 0 && exec.CurrentExecutable.EstimatedDuration > 0 && exec.CurrentExecutable.Timestamp > 0 {
					progress = exec.CurrentExecutable.GetProgress()
				}

				running = append(running, RunningBuildInfo{
					JobName:  jobName,
					BuildNum: exec.CurrentExecutable.Number,
					NodeName: node.DisplayName,
					URL:      exec.CurrentExecutable.URL,
					Progress: progress,
				})
			}
		}
	}
	return running
}

func (m *DashboardModel) processJobs(jobs []models.Job) {
	// Sort by last build timestamp (most recent first)
	models.SortJobsByLastBuild(jobs)

	// Build list of recent builds (all builds, not just failed)
	m.recentBuilds = nil
	for _, job := range jobs {
		if job.LastBuild != nil {
			result := job.LastBuild.Result
			if result == "" {
				// Check if it's running
				if job.IsRunning() {
					result = "RUNNING"
				}
			}
			m.recentBuilds = append(m.recentBuilds, RecentBuildInfo{
				JobName:   job.Name,
				BuildNum:  job.LastBuild.Number,
				Result:    result,
				Color:     job.Color,
				Timestamp: time.UnixMilli(job.LastBuild.Timestamp),
				Duration:  time.Duration(job.LastBuild.Duration) * time.Millisecond,
				URL:       job.URL,
			})
		}
	}
}

func (m *DashboardModel) fetchRootInfo() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		info, err := m.client.GetRootInfo(ctx)
		return DashboardDataMsg{RootInfo: info, Error: err}
	}
}

func (m *DashboardModel) fetchNodes() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		nodes, err := m.client.GetNodes(ctx)
		return DashboardDataMsg{Nodes: nodes, Error: err}
	}
}

func (m *DashboardModel) fetchQueue() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		queue, err := m.client.GetQueue(ctx)
		return DashboardDataMsg{Queue: queue, Error: err}
	}
}

func (m *DashboardModel) fetchJobs() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		jobs, err := m.client.GetAllJobs(ctx)
		return DashboardDataMsg{Jobs: jobs, Error: err}
	}
}

// DashboardDataMsg carries dashboard data updates
type DashboardDataMsg struct {
	RootInfo *models.RootInfo
	Nodes    []models.Node
	Queue    *models.Queue
	Jobs     []models.Job
	Error    error
}
