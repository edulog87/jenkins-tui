// Package app contains the main application model and logic for the TUI.
package app

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elogrono/jenkins-tui/internal/config"
	"github.com/elogrono/jenkins-tui/internal/jenkins"
	"github.com/elogrono/jenkins-tui/internal/logger"
	"github.com/elogrono/jenkins-tui/internal/ui/theme"
)

// TabID represents the different tabs in the application
type TabID int

const (
	TabDashboard TabID = iota
	TabViews
	TabBuilds
)

// AppState represents the current state of the application
type AppState int

const (
	StateSetup   AppState = iota // First-run configuration
	StateLoading                 // Loading data
	StateReady                   // Normal operation
	StateError                   // Error state
)

// Model is the main application model
type Model struct {
	// Configuration
	config *config.Config

	// Jenkins client
	client *jenkins.Client

	// Application state
	state     AppState
	activeTab TabID

	// UI state
	width  int
	height int

	// Loading spinner
	spinner spinner.Model

	// Error message (if any)
	lastError error

	// Setup wizard state (for first-run)
	setupModel *SetupModel

	// Tab models
	dashboardModel *DashboardModel
	viewsModel     *ViewsModel
	buildsModel    *BuildsModel

	// Help visibility
	showHelp bool

	// Auto-refresh
	autoRefreshEnabled  bool
	autoRefreshInterval time.Duration
}

// NewModel creates a new application model
func NewModel(cfg *config.Config) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.SpinnerStyle()

	// Default auto-refresh interval from config
	refreshInterval := time.Duration(cfg.Profile.AutoRefreshSeconds) * time.Second
	if refreshInterval < 5*time.Second {
		refreshInterval = 10 * time.Second
	}

	m := &Model{
		config:              cfg,
		state:               StateLoading,
		activeTab:           TabDashboard,
		spinner:             s,
		autoRefreshEnabled:  true,
		autoRefreshInterval: refreshInterval,
	}

	// Check if we need to run setup wizard
	if !cfg.IsConfigured() {
		logger.Info("Config not complete, entering setup wizard")
		m.state = StateSetup
		m.setupModel = NewSetupModel()
	} else {
		logger.Info("Config is complete, will initialize client")
	}

	return m
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	logger.Debug("Model.Init called", "state", m.state)

	cmds := []tea.Cmd{
		m.spinner.Tick,
	}

	if m.state == StateSetup {
		logger.Debug("Init: entering setup mode")
		cmds = append(cmds, m.setupModel.Init())
	} else if m.state == StateLoading {
		logger.Debug("Init: starting client initialization")
		cmds = append(cmds, m.initializeClient())
	}

	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key handling
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state != StateSetup {
				return m, tea.Quit
			}
		case "r":
			// Retry connection on error state
			if m.state == StateError {
				logger.Info("Retrying connection...")
				m.state = StateLoading
				m.lastError = nil
				return m, m.initializeClient()
			}
		case "?":
			if m.state == StateReady {
				m.showHelp = !m.showHelp
				return m, nil
			}
		case "ctrl+r":
			// Toggle auto-refresh
			if m.state == StateReady {
				m.autoRefreshEnabled = !m.autoRefreshEnabled
				if m.autoRefreshEnabled {
					logger.Info("Auto-refresh enabled")
					return m, m.scheduleAutoRefresh()
				}
				logger.Info("Auto-refresh disabled")
				return m, nil
			}
		case "tab":
			if m.state == StateReady && !m.showHelp {
				m.activeTab = (m.activeTab + 1) % 3
				return m, m.loadTabData()
			}
		case "shift+tab":
			if m.state == StateReady && !m.showHelp {
				m.activeTab = (m.activeTab + 2) % 3
				return m, m.loadTabData()
			}
		case "1":
			if m.state == StateReady {
				m.activeTab = TabDashboard
				return m, m.loadTabData()
			}
		case "2":
			if m.state == StateReady {
				m.activeTab = TabViews
				return m, m.loadTabData()
			}
		case "3":
			if m.state == StateReady {
				m.activeTab = TabBuilds
				return m, m.loadTabData()
			}
		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propagate to sub-models
		if m.setupModel != nil {
			m.setupModel.width = msg.Width
			m.setupModel.height = msg.Height
		}
		if m.dashboardModel != nil {
			m.dashboardModel.SetSize(msg.Width, msg.Height)
		}
		if m.viewsModel != nil {
			m.viewsModel.SetSize(msg.Width, msg.Height)
		}
		if m.buildsModel != nil {
			m.buildsModel.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case SetupCompleteMsg:
		// Setup wizard completed
		logger.Info("Setup complete, saving config")
		m.config = msg.Config
		if err := m.config.Save(); err != nil {
			logger.Error("Failed to save config", "error", err)
			m.lastError = err
			m.state = StateError
			return m, nil
		}
		logger.Info("Config saved, initializing client")
		m.state = StateLoading
		m.setupModel = nil
		return m, m.initializeClient()

	case ClientReadyMsg:
		logger.Info("Client is ready, transitioning to StateReady")
		m.client = msg.Client
		m.state = StateReady
		m.initTabModels()
		// Start auto-refresh timer
		return m, tea.Batch(m.loadTabData(), m.scheduleAutoRefresh())

	case ClientErrorMsg:
		logger.Error("Client error received", "error", msg.Error)
		m.lastError = msg.Error
		m.state = StateError
		return m, nil

	case AutoRefreshTickMsg:
		// Auto-refresh current tab data
		if m.state == StateReady && m.autoRefreshEnabled {
			logger.Debug("Auto-refresh triggered")
			return m, tea.Batch(m.loadTabData(), m.scheduleAutoRefresh())
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Delegate to appropriate sub-model based on state
	switch m.state {
	case StateSetup:
		if m.setupModel != nil {
			newSetup, cmd := m.setupModel.Update(msg)
			m.setupModel = newSetup.(*SetupModel)
			cmds = append(cmds, cmd)
		}
	case StateReady:
		cmd := m.updateActiveTab(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m *Model) View() string {
	switch m.state {
	case StateSetup:
		if m.setupModel != nil {
			return m.setupModel.View()
		}
	case StateLoading:
		return m.viewLoading()
	case StateError:
		return m.viewError()
	case StateReady:
		if m.showHelp {
			return m.viewHelp()
		}
		return m.viewMain()
	}
	return ""
}

// initializeClient creates the Jenkins client
func (m *Model) initializeClient() tea.Cmd {
	logger.Info("initializeClient command starting")
	return func() tea.Msg {
		logger.Debug("Inside initializeClient goroutine")
		client, err := jenkins.NewClient(m.config)
		if err != nil {
			logger.Error("Failed to create client", "error", err)
			return ClientErrorMsg{Error: err}
		}

		logger.Info("Client created, testing connection...")
		// Test connection
		if err := client.TestConnection(); err != nil {
			logger.Error("Connection test failed", "error", err)
			return ClientErrorMsg{Error: err}
		}

		logger.Info("Connection test passed, returning ClientReadyMsg")
		return ClientReadyMsg{Client: client}
	}
}

// initTabModels initializes the tab models
func (m *Model) initTabModels() {
	m.dashboardModel = NewDashboardModel(m.client, m.width, m.height)
	m.viewsModel = NewViewsModel(m.client, m.width, m.height)
	m.buildsModel = NewBuildsModel(m.client, m.width, m.height)
}

// loadTabData loads data for the current tab
func (m *Model) loadTabData() tea.Cmd {
	switch m.activeTab {
	case TabDashboard:
		if m.dashboardModel != nil {
			return m.dashboardModel.LoadData()
		}
	case TabViews:
		if m.viewsModel != nil {
			return m.viewsModel.LoadData()
		}
	case TabBuilds:
		if m.buildsModel != nil {
			return m.buildsModel.LoadData()
		}
	}
	return nil
}

// updateActiveTab delegates updates to the active tab
func (m *Model) updateActiveTab(msg tea.Msg) tea.Cmd {
	switch m.activeTab {
	case TabDashboard:
		if m.dashboardModel != nil {
			return m.dashboardModel.Update(msg)
		}
	case TabViews:
		if m.viewsModel != nil {
			return m.viewsModel.Update(msg)
		}
	case TabBuilds:
		if m.buildsModel != nil {
			return m.buildsModel.Update(msg)
		}
	}
	return nil
}

// Message types
type SetupCompleteMsg struct {
	Config *config.Config
}

type ClientReadyMsg struct {
	Client *jenkins.Client
}

type ClientErrorMsg struct {
	Error error
}

// AutoRefreshTickMsg is sent periodically for auto-refresh
type AutoRefreshTickMsg struct{}

// scheduleAutoRefresh returns a command that triggers auto-refresh after the interval
func (m *Model) scheduleAutoRefresh() tea.Cmd {
	if !m.autoRefreshEnabled || m.autoRefreshInterval <= 0 {
		return nil
	}
	return tea.Tick(m.autoRefreshInterval, func(t time.Time) tea.Msg {
		return AutoRefreshTickMsg{}
	})
}
