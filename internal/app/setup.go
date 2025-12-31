package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elogrono/jenkins-tui/internal/config"
	"github.com/elogrono/jenkins-tui/internal/jenkins"
	"github.com/elogrono/jenkins-tui/internal/ui/theme"
)

// SetupField represents which field is currently focused
type SetupField int

const (
	FieldURL SetupField = iota
	FieldUsername
	FieldToken
	FieldSubmit
)

// SetupModel handles the first-run configuration wizard
type SetupModel struct {
	width  int
	height int

	urlInput      textinput.Model
	usernameInput textinput.Model
	tokenInput    textinput.Model

	focusedField SetupField
	err          error
	testing      bool
	testResult   string
}

// NewSetupModel creates a new setup wizard model
func NewSetupModel() *SetupModel {
	urlInput := textinput.New()
	urlInput.Placeholder = "https://jenkins.example.com"
	urlInput.Focus()
	urlInput.CharLimit = 256
	urlInput.Width = 50

	usernameInput := textinput.New()
	usernameInput.Placeholder = "admin"
	usernameInput.CharLimit = 64
	usernameInput.Width = 50

	tokenInput := textinput.New()
	tokenInput.Placeholder = "API Token"
	tokenInput.EchoMode = textinput.EchoPassword
	tokenInput.EchoCharacter = '*'
	tokenInput.CharLimit = 128
	tokenInput.Width = 50

	return &SetupModel{
		urlInput:      urlInput,
		usernameInput: usernameInput,
		tokenInput:    tokenInput,
		focusedField:  FieldURL,
	}
}

// Init implements tea.Model
func (m *SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "tab", "down":
			m.nextField()
			return m, nil

		case "shift+tab", "up":
			m.prevField()
			return m, nil

		case "enter":
			if m.focusedField == FieldSubmit {
				return m, m.testAndSave()
			}
			m.nextField()
			return m, nil
		}

	case testConnectionResult:
		m.testing = false
		if msg.err != nil {
			m.err = msg.err
			m.testResult = fmt.Sprintf("Connection failed: %v", msg.err)
		} else {
			m.testResult = "Connection successful!"
			// Return the completed config
			cfg := config.DefaultConfig()
			cfg.Profile.BaseURL = m.urlInput.Value()
			cfg.Profile.Username = m.usernameInput.Value()
			cfg.Profile.APIToken = m.tokenInput.Value()
			return m, func() tea.Msg {
				return SetupCompleteMsg{Config: cfg}
			}
		}
		return m, nil
	}

	// Update the focused input
	var cmd tea.Cmd
	switch m.focusedField {
	case FieldURL:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case FieldUsername:
		m.usernameInput, cmd = m.usernameInput.Update(msg)
	case FieldToken:
		m.tokenInput, cmd = m.tokenInput.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m *SetupModel) View() string {
	var b strings.Builder

	// Title
	title := theme.TitleStyle.Render("Jenkins TUI - Initial Setup")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Instructions
	instructions := theme.MutedStyle.Render("Enter your Jenkins server details to get started.\nYou can find your API token in Jenkins > User > Configure > API Token")
	b.WriteString(instructions)
	b.WriteString("\n\n")

	// URL field
	b.WriteString(m.renderField("Jenkins URL:", m.urlInput.View(), m.focusedField == FieldURL))
	b.WriteString("\n\n")

	// Username field
	b.WriteString(m.renderField("Username:", m.usernameInput.View(), m.focusedField == FieldUsername))
	b.WriteString("\n\n")

	// Token field
	b.WriteString(m.renderField("API Token:", m.tokenInput.View(), m.focusedField == FieldToken))
	b.WriteString("\n\n")

	// Submit button
	buttonStyle := theme.ButtonStyle
	if m.focusedField == FieldSubmit {
		buttonStyle = theme.ButtonFocusedStyle
	}
	button := buttonStyle.Render("[ Test Connection & Save ]")
	b.WriteString(button)
	b.WriteString("\n\n")

	// Status messages
	if m.testing {
		b.WriteString(theme.MutedStyle.Render("Testing connection..."))
	} else if m.testResult != "" {
		if m.err != nil {
			b.WriteString(theme.ErrorStyle.Render(m.testResult))
		} else {
			b.WriteString(theme.SuccessStyle.Render(m.testResult))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(theme.MutedStyle.Render("Tab/Shift+Tab: Navigate | Enter: Submit | Ctrl+C: Quit"))

	// Center the content
	content := b.String()
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(2, 4).
		Width(70)

	box := boxStyle.Render(content)

	// Center vertically and horizontally
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

func (m *SetupModel) renderField(label, input string, focused bool) string {
	labelStyle := theme.InputLabelStyle
	if focused {
		labelStyle = labelStyle.Foreground(theme.Primary)
	}
	return labelStyle.Render(label) + "\n" + input
}

func (m *SetupModel) nextField() {
	m.focusedField = (m.focusedField + 1) % 4
	m.updateFocus()
}

func (m *SetupModel) prevField() {
	m.focusedField = (m.focusedField + 3) % 4
	m.updateFocus()
}

func (m *SetupModel) updateFocus() {
	m.urlInput.Blur()
	m.usernameInput.Blur()
	m.tokenInput.Blur()

	switch m.focusedField {
	case FieldURL:
		m.urlInput.Focus()
	case FieldUsername:
		m.usernameInput.Focus()
	case FieldToken:
		m.tokenInput.Focus()
	}
}

func (m *SetupModel) testAndSave() tea.Cmd {
	// Validate inputs
	url := strings.TrimSpace(m.urlInput.Value())
	username := strings.TrimSpace(m.usernameInput.Value())
	token := strings.TrimSpace(m.tokenInput.Value())

	if url == "" || username == "" || token == "" {
		m.err = fmt.Errorf("all fields are required")
		m.testResult = "All fields are required"
		return nil
	}

	m.testing = true
	m.err = nil
	m.testResult = ""

	return func() tea.Msg {
		cfg := config.DefaultConfig()
		cfg.Profile.BaseURL = url
		cfg.Profile.Username = username
		cfg.Profile.APIToken = token

		client, err := jenkins.NewClient(cfg)
		if err != nil {
			return testConnectionResult{err: err}
		}

		err = client.TestConnection()
		return testConnectionResult{err: err}
	}
}

type testConnectionResult struct {
	err error
}
