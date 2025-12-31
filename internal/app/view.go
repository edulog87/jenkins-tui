package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/elogrono/jenkins-tui/internal/logger"
	"github.com/elogrono/jenkins-tui/internal/ui/theme"
)

// viewLoading renders the loading state
func (m *Model) viewLoading() string {
	jenkinsURL := ""
	if m.config != nil && m.config.Profile.BaseURL != "" {
		jenkinsURL = m.config.Profile.BaseURL
	}

	lines := []string{
		fmt.Sprintf("%s Connecting to Jenkins...", m.spinner.View()),
		"",
	}

	if jenkinsURL != "" {
		lines = append(lines, theme.MutedStyle.Render(fmt.Sprintf("URL: %s", jenkinsURL)))
	}

	lines = append(lines,
		"",
		theme.MutedStyle.Render(fmt.Sprintf("Log: %s", logger.LogFilePath())),
		theme.MutedStyle.Render("(tail -f to monitor progress)"),
	)

	content := lipgloss.JoinVertical(lipgloss.Center, lines...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(2, 4).
		Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// viewError renders the error state
func (m *Model) viewError() string {
	errMsg := "Unknown error"
	if m.lastError != nil {
		errMsg = m.lastError.Error()
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		theme.ErrorStyle.Render("Connection Error"),
		"",
		theme.MutedStyle.Render(errMsg),
		"",
		theme.MutedStyle.Render("Press 'q' to quit or 'r' to retry"),
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Error).
		Padding(2, 4).
		Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// viewHelp renders the help panel
func (m *Model) viewHelp() string {
	helpContent := `
GLOBAL KEYS
  Tab/Shift+Tab    Navigate tabs
  1/2/3            Jump to tab
  ?                Toggle help
  q/Ctrl+C         Quit

DASHBOARD
  r                Refresh data
  Enter            View details
  f                Quick filters

VIEWS
  /                Search
  Enter            Open view/job
  Backspace        Clear filter
  Esc              Go back
  r                Refresh

BUILDS
  /                Search
  Enter            View details
  l                View logs
  PgUp/PgDn        Navigate pages
  Esc              Go back
  r                Refresh

LOG VIEWER
  /                Search in log
  s                Toggle follow
  g/G              Top/Bottom
  PgUp/PgDn        Scroll
  Esc              Go back
`

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(1, 2).
		Width(50).
		Render(helpContent)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

// viewMain renders the main application view
func (m *Model) viewMain() string {
	// Render tabs
	tabs := m.renderTabs()

	// Render content based on active tab
	var content string
	switch m.activeTab {
	case TabDashboard:
		if m.dashboardModel != nil {
			content = m.dashboardModel.View()
		}
	case TabViews:
		if m.viewsModel != nil {
			content = m.viewsModel.View()
		}
	case TabBuilds:
		if m.buildsModel != nil {
			content = m.buildsModel.View()
		}
	}

	// Render status bar
	statusBar := m.renderStatusBar()

	// Compose layout
	return lipgloss.JoinVertical(lipgloss.Left,
		tabs,
		content,
		statusBar,
	)
}

// renderTabs renders the tab bar
func (m *Model) renderTabs() string {
	tabNames := []string{"Dashboard", "Views", "Builds"}

	var tabs []string
	for i, name := range tabNames {
		style := theme.TabStyle(i == int(m.activeTab))
		tabs = append(tabs, style.Render(fmt.Sprintf("[%d] %s", i+1, name)))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	return lipgloss.NewStyle().
		BorderBottom(true).
		BorderForeground(theme.Border).
		Width(m.width).
		Render(tabBar)
}

// renderStatusBar renders the status bar
func (m *Model) renderStatusBar() string {
	left := theme.MutedStyle.Render("Jenkins TUI")

	right := theme.MutedStyle.Render("? Help | q Quit")

	// Calculate padding
	padding := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if padding < 0 {
		padding = 0
	}

	return theme.StatusBarStyle.
		Width(m.width).
		Render(fmt.Sprintf("%s%*s%s", left, padding, "", right))
}
