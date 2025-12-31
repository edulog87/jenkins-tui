// Package theme provides consistent styling for the TUI.
package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PROFESSIONAL COLOR PALETTE
// ═══════════════════════════════════════════════════════════════════════════════

// Base Colors
var (
	Primary       = lipgloss.Color("#8B5CF6") // Vivid Purple
	Secondary     = lipgloss.Color("#06B6D4") // Cyan
	Accent        = lipgloss.Color("#F59E0B") // Amber
	Highlight     = lipgloss.Color("#EC4899") // Pink
	Background    = lipgloss.Color("#0F172A") // Deep slate
	Surface       = lipgloss.Color("#1E293B") // Slate surface
	SurfaceAlt    = lipgloss.Color("#334155") // Lighter slate
	Foreground    = lipgloss.Color("#F8FAFC") // Almost white
	ForegroundDim = lipgloss.Color("#94A3B8") // Muted text
	Border        = lipgloss.Color("#475569") // Slate border
	BorderFocus   = lipgloss.Color("#8B5CF6") // Purple focus
)

// Status Colors
var (
	Success  = lipgloss.Color("#22C55E") // Bright Green
	Error    = lipgloss.Color("#EF4444") // Red
	Warning  = lipgloss.Color("#F59E0B") // Amber
	Info     = lipgloss.Color("#3B82F6") // Blue
	Muted    = lipgloss.Color("#64748B") // Slate gray
	Running  = lipgloss.Color("#06B6D4") // Cyan
	Pending  = lipgloss.Color("#A855F7") // Purple
	Disabled = lipgloss.Color("#475569") // Dark slate
)

// Jenkins Build Status Colors
var (
	BuildSuccess  = lipgloss.Color("#22C55E") // Green
	BuildFailure  = lipgloss.Color("#EF4444") // Red
	BuildUnstable = lipgloss.Color("#FBBF24") // Yellow
	BuildAborted  = lipgloss.Color("#6B7280") // Gray
	BuildRunning  = lipgloss.Color("#06B6D4") // Cyan animated
	BuildNotBuilt = lipgloss.Color("#64748B") // Slate
	BuildDisabled = lipgloss.Color("#374151") // Dark gray
)

// ═══════════════════════════════════════════════════════════════════════════════
// UNICODE ICONS
// ═══════════════════════════════════════════════════════════════════════════════

// Status Icons
const (
	IconSuccess  = "✓"
	IconFailure  = "✗"
	IconWarning  = "⚠"
	IconRunning  = "●"
	IconPending  = "○"
	IconAborted  = "⊘"
	IconNotBuilt = "○"
	IconDisabled = "⊝"
	IconUnknown  = "?"

	// Navigation
	IconArrowRight = "→"
	IconArrowLeft  = "←"
	IconArrowUp    = "↑"
	IconArrowDown  = "↓"
	IconExpand     = "▶"
	IconCollapse   = "▼"
	IconFolder     = "📁"
	IconFolderOpen = "📂"
	IconFile       = "📄"

	// UI Elements
	IconSearch    = "🔍"
	IconFilter    = "⚙"
	IconRefresh   = "⟳"
	IconClock     = "⏱"
	IconCalendar  = "📅"
	IconUser      = "👤"
	IconServer    = "🖥"
	IconBuild     = "🔨"
	IconQueue     = "📋"
	IconLog       = "📜"
	IconArtifact  = "📦"
	IconBranch    = "⎇"
	IconCommit    = "●"
	IconLink      = "🔗"
	IconStar      = "★"
	IconStarEmpty = "☆"

	// Progress
	IconSpinner1 = "⠋"
	IconSpinner2 = "⠙"
	IconSpinner3 = "⠹"
	IconSpinner4 = "⠸"
	IconSpinner5 = "⠼"
	IconSpinner6 = "⠴"
	IconSpinner7 = "⠦"
	IconSpinner8 = "⠧"

	// Health
	IconHealthFull  = "████"
	IconHealth75    = "███░"
	IconHealth50    = "██░░"
	IconHealth25    = "█░░░"
	IconHealthEmpty = "░░░░"

	// Stage status
	IconStagePass = "━━━"
	IconStageFail = "━━━"
	IconStageSkip = "- - -"
	IconStageRun  = "━ ▶ ━"
)

// ═══════════════════════════════════════════════════════════════════════════════
// TYPOGRAPHY STYLES
// ═══════════════════════════════════════════════════════════════════════════════

var (
	// Titles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Foreground).
			Background(Primary).
			Padding(0, 2).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Secondary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(Border).
				PaddingBottom(0)

	// Text
	BaseStyle = lipgloss.NewStyle().
			Foreground(Foreground)

	DimStyle = lipgloss.NewStyle().
			Foreground(ForegroundDim)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	AccentStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	// Status text
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Info)

	RunningStyle = lipgloss.NewStyle().
			Foreground(Running).
			Bold(true)

	PrimaryStyle = lipgloss.NewStyle().
			Foreground(Primary)
)

// ═══════════════════════════════════════════════════════════════════════════════
// CONTAINER STYLES
// ═══════════════════════════════════════════════════════════════════════════════

var (
	// Boxes
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2)

	FocusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderFocus).
			Padding(1, 2)

	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	// Cards
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SurfaceAlt).
			Background(Surface).
			Padding(1, 2)

	// KPI Card
	KPICardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 2).
			Align(lipgloss.Center)
)

// ═══════════════════════════════════════════════════════════════════════════════
// TAB STYLES
// ═══════════════════════════════════════════════════════════════════════════════

// TabStyle returns the style for a tab based on active state
func TabStyle(active bool) lipgloss.Style {
	base := lipgloss.NewStyle().
		Padding(0, 3).
		MarginRight(1)

	if active {
		return base.
			Bold(true).
			Foreground(Background).
			Background(Primary)
	}
	return base.
		Foreground(ForegroundDim).
		Background(Surface)
}

// TabBarStyle for the tab container
var TabBarStyle = lipgloss.NewStyle().
	BorderBottom(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(Border).
	PaddingBottom(0).
	MarginBottom(1)

// ═══════════════════════════════════════════════════════════════════════════════
// STATUS BAR STYLES
// ═══════════════════════════════════════════════════════════════════════════════

var (
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ForegroundDim).
			Background(Surface).
			Padding(0, 1)

	StatusBarKeyStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Background(Surface).
				Bold(true).
				Padding(0, 1)

	StatusBarDescStyle = lipgloss.NewStyle().
				Foreground(ForegroundDim).
				Background(Surface)
)

// ═══════════════════════════════════════════════════════════════════════════════
// BUTTON STYLES
// ═══════════════════════════════════════════════════════════════════════════════

var (
	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Background(SurfaceAlt).
			Foreground(Foreground)

	ButtonFocusedStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Background(Primary).
				Foreground(Background).
				Bold(true)

	ButtonPrimaryStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Background(Primary).
				Foreground(Background).
				Bold(true)

	ButtonDangerStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Background(Error).
				Foreground(Foreground).
				Bold(true)
)

// ═══════════════════════════════════════════════════════════════════════════════
// TABLE STYLES
// ═══════════════════════════════════════════════════════════════════════════════

var (
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Secondary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(Border).
				Padding(0, 1)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(Foreground).
			Padding(0, 1)

	TableRowAltStyle = lipgloss.NewStyle().
				Foreground(Foreground).
				Background(Surface).
				Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(Background).
				Background(Primary).
				Bold(true).
				Padding(0, 1)

	TableCellMutedStyle = lipgloss.NewStyle().
				Foreground(Muted).
				Padding(0, 1)
)

// ═══════════════════════════════════════════════════════════════════════════════
// INPUT STYLES
// ═══════════════════════════════════════════════════════════════════════════════

var (
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(0, 1)

	InputLabelStyle = lipgloss.NewStyle().
			Foreground(Foreground).
			Bold(true).
			MarginBottom(1)

	InputPlaceholderStyle = lipgloss.NewStyle().
				Foreground(Muted)

	SearchBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1).
			MarginBottom(1)
)

// ═══════════════════════════════════════════════════════════════════════════════
// HELP STYLES
// ═══════════════════════════════════════════════════════════════════════════════

var (
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Padding(0, 1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ForegroundDim)

	HelpSeparatorStyle = lipgloss.NewStyle().
				Foreground(Border)
)

// ═══════════════════════════════════════════════════════════════════════════════
// SPINNER
// ═══════════════════════════════════════════════════════════════════════════════

// SpinnerStyle returns the style for the loading spinner
func SpinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Running)
}

// ═══════════════════════════════════════════════════════════════════════════════
// BUILD STATUS HELPERS
// ═══════════════════════════════════════════════════════════════════════════════

// BuildStatusStyle returns the appropriate style for a Jenkins build status
func BuildStatusStyle(color string) lipgloss.Style {
	base := lipgloss.NewStyle().Bold(true)
	switch color {
	case "blue", "blue_anime":
		return base.Foreground(BuildSuccess)
	case "red", "red_anime":
		return base.Foreground(BuildFailure)
	case "yellow", "yellow_anime":
		return base.Foreground(BuildUnstable)
	case "aborted", "aborted_anime":
		return base.Foreground(BuildAborted)
	case "notbuilt", "notbuilt_anime":
		return base.Foreground(BuildNotBuilt)
	case "disabled":
		return base.Foreground(BuildDisabled)
	default:
		return base.Foreground(Muted)
	}
}

// BuildStatusIcon returns a styled icon for the build status
func BuildStatusIcon(color string) string {
	switch color {
	case "blue":
		return lipgloss.NewStyle().Foreground(BuildSuccess).Bold(true).Render(IconSuccess)
	case "blue_anime":
		return lipgloss.NewStyle().Foreground(BuildRunning).Bold(true).Render(IconRunning)
	case "red":
		return lipgloss.NewStyle().Foreground(BuildFailure).Bold(true).Render(IconFailure)
	case "red_anime":
		return lipgloss.NewStyle().Foreground(BuildRunning).Bold(true).Render(IconRunning)
	case "yellow":
		return lipgloss.NewStyle().Foreground(BuildUnstable).Bold(true).Render(IconWarning)
	case "yellow_anime":
		return lipgloss.NewStyle().Foreground(BuildRunning).Bold(true).Render(IconRunning)
	case "aborted":
		return lipgloss.NewStyle().Foreground(BuildAborted).Render(IconAborted)
	case "aborted_anime":
		return lipgloss.NewStyle().Foreground(BuildRunning).Bold(true).Render(IconRunning)
	case "notbuilt":
		return lipgloss.NewStyle().Foreground(BuildNotBuilt).Render(IconNotBuilt)
	case "disabled":
		return lipgloss.NewStyle().Foreground(BuildDisabled).Render(IconDisabled)
	default:
		return lipgloss.NewStyle().Foreground(Muted).Render(IconUnknown)
	}
}

// BuildResultStyle returns the style based on result string
func BuildResultStyle(result string) lipgloss.Style {
	base := lipgloss.NewStyle().Bold(true)
	switch result {
	case "SUCCESS":
		return base.Foreground(BuildSuccess)
	case "FAILURE":
		return base.Foreground(BuildFailure)
	case "UNSTABLE":
		return base.Foreground(BuildUnstable)
	case "ABORTED":
		return base.Foreground(BuildAborted)
	case "RUNNING", "BUILDING":
		return base.Foreground(BuildRunning)
	case "NOT_BUILT":
		return base.Foreground(BuildNotBuilt)
	default:
		return base.Foreground(Muted)
	}
}

// BuildResultIcon returns icon based on result
func BuildResultIcon(result string) string {
	switch result {
	case "SUCCESS":
		return lipgloss.NewStyle().Foreground(BuildSuccess).Bold(true).Render(IconSuccess)
	case "FAILURE":
		return lipgloss.NewStyle().Foreground(BuildFailure).Bold(true).Render(IconFailure)
	case "UNSTABLE":
		return lipgloss.NewStyle().Foreground(BuildUnstable).Bold(true).Render(IconWarning)
	case "ABORTED":
		return lipgloss.NewStyle().Foreground(BuildAborted).Render(IconAborted)
	case "RUNNING", "BUILDING":
		return lipgloss.NewStyle().Foreground(BuildRunning).Bold(true).Render(IconRunning)
	default:
		return lipgloss.NewStyle().Foreground(Muted).Render(IconPending)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// HEALTH BAR
// ═══════════════════════════════════════════════════════════════════════════════

// HealthBar returns a visual health bar
func HealthBar(score int) string {
	var bar string
	var color lipgloss.Color

	switch {
	case score >= 80:
		bar = "████"
		color = BuildSuccess
	case score >= 60:
		bar = "███░"
		color = lipgloss.Color("#84CC16") // Lime
	case score >= 40:
		bar = "██░░"
		color = BuildUnstable
	case score >= 20:
		bar = "█░░░"
		color = lipgloss.Color("#F97316") // Orange
	default:
		bar = "░░░░"
		color = BuildFailure
	}

	return lipgloss.NewStyle().Foreground(color).Render(bar)
}

// ═══════════════════════════════════════════════════════════════════════════════
// PROGRESS BAR
// ═══════════════════════════════════════════════════════════════════════════════

// ProgressBar returns a visual progress bar
func ProgressBar(percent int, width int) string {
	if width < 5 {
		width = 10
	}
	filled := (percent * width) / 100
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return lipgloss.NewStyle().Foreground(Primary).Render(bar)
}

// ═══════════════════════════════════════════════════════════════════════════════
// SHORTKEY STYLES
// ═══════════════════════════════════════════════════════════════════════════════

// FormatShortkey formats a keyboard shortcut
func FormatShortkey(key, description string) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Background(Surface).
		Padding(0, 1)

	descStyle := lipgloss.NewStyle().
		Foreground(ForegroundDim)

	return keyStyle.Render(key) + " " + descStyle.Render(description)
}

// ShortkeysBar creates a horizontal bar of shortcuts
func ShortkeysBar(shortcuts map[string]string, width int) string {
	var items []string
	for key, desc := range shortcuts {
		items = append(items, FormatShortkey(key, desc))
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Left, items...)
	return lipgloss.NewStyle().
		Background(Surface).
		Width(width).
		Padding(0, 1).
		Render(bar)
}

// ═══════════════════════════════════════════════════════════════════════════════
// BADGE STYLES
// ═══════════════════════════════════════════════════════════════════════════════

// Badge creates a styled badge
func Badge(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(Background).
		Background(color).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// StatusBadge creates a status badge
func StatusBadge(status string) string {
	var color lipgloss.Color
	switch status {
	case "SUCCESS":
		color = BuildSuccess
	case "FAILURE":
		color = BuildFailure
	case "UNSTABLE":
		color = BuildUnstable
	case "RUNNING", "BUILDING":
		color = BuildRunning
	case "ABORTED":
		color = BuildAborted
	default:
		color = Muted
	}
	return Badge(status, color)
}
