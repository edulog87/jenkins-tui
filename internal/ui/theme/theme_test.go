package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBuildStatusIcon(t *testing.T) {
	// The icons now include ANSI styling, so we check for the icon character
	tests := []struct {
		color    string
		contains string
	}{
		{"blue", IconSuccess},
		{"blue_anime", IconRunning},
		{"red", IconFailure},
		{"red_anime", IconRunning},
		{"yellow", IconWarning},
		{"yellow_anime", IconRunning},
		{"aborted", IconAborted},
		{"aborted_anime", IconRunning},
		{"notbuilt", IconNotBuilt},
		{"disabled", IconDisabled},
		{"unknown", IconUnknown},
	}

	for _, tt := range tests {
		result := BuildStatusIcon(tt.color)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("BuildStatusIcon(%s) should contain %s, got %s", tt.color, tt.contains, result)
		}
	}
}

func TestBuildStatusStyle(t *testing.T) {
	colors := []string{
		"blue", "blue_anime",
		"red", "red_anime",
		"yellow", "yellow_anime",
		"aborted", "aborted_anime",
		"notbuilt", "notbuilt_anime",
		"unknown",
	}

	for _, color := range colors {
		style := BuildStatusStyle(color)
		// Just verify it returns a valid style without panicking
		_ = style.Render("test")
	}
}

func TestBuildResultStyle(t *testing.T) {
	results := []string{
		"SUCCESS", "FAILURE", "UNSTABLE",
		"ABORTED", "RUNNING", "NOT_BUILT",
		"UNKNOWN",
	}

	for _, result := range results {
		style := BuildResultStyle(result)
		rendered := style.Render("test")
		if rendered == "" {
			t.Errorf("BuildResultStyle(%s) rendered empty string", result)
		}
	}
}

func TestBuildResultIcon(t *testing.T) {
	results := []string{
		"SUCCESS", "FAILURE", "UNSTABLE",
		"ABORTED", "RUNNING", "PENDING",
	}

	for _, result := range results {
		icon := BuildResultIcon(result)
		if icon == "" {
			t.Errorf("BuildResultIcon(%s) returned empty string", result)
		}
	}
}

func TestTabStyle(t *testing.T) {
	activeStyle := TabStyle(true)
	inactiveStyle := TabStyle(false)

	// Verify both return valid styles
	activeText := activeStyle.Render("Tab")
	inactiveText := inactiveStyle.Render("Tab")

	if activeText == "" || inactiveText == "" {
		t.Error("TabStyle returned empty render")
	}
}

func TestSpinnerStyle(t *testing.T) {
	style := SpinnerStyle()
	// Verify it returns a valid style
	result := style.Render("*")
	if result == "" {
		t.Error("SpinnerStyle rendered empty string")
	}
}

func TestHealthBar(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{100, "████"},
		{80, "████"},
		{60, "███░"},
		{40, "██░░"},
		{20, "█░░░"},
		{0, "░░░░"},
	}

	for _, tt := range tests {
		result := HealthBar(tt.score)
		if !strings.Contains(result, tt.expected) {
			t.Errorf("HealthBar(%d) should contain %s", tt.score, tt.expected)
		}
	}
}

func TestProgressBar(t *testing.T) {
	tests := []struct {
		percent int
		width   int
	}{
		{0, 10},
		{50, 10},
		{100, 10},
		{150, 10}, // Should cap at 100
	}

	for _, tt := range tests {
		result := ProgressBar(tt.percent, tt.width)
		if result == "" {
			t.Errorf("ProgressBar(%d, %d) returned empty string", tt.percent, tt.width)
		}
	}
}

func TestFormatShortkey(t *testing.T) {
	result := FormatShortkey("Enter", "Select")
	if result == "" {
		t.Error("FormatShortkey returned empty string")
	}
	if !strings.Contains(result, "Enter") || !strings.Contains(result, "Select") {
		t.Errorf("FormatShortkey should contain key and description: %s", result)
	}
}

func TestBadge(t *testing.T) {
	result := Badge("TEST", Success)
	if result == "" {
		t.Error("Badge returned empty string")
	}
	if !strings.Contains(result, "TEST") {
		t.Error("Badge should contain the text")
	}
}

func TestStatusBadge(t *testing.T) {
	statuses := []string{"SUCCESS", "FAILURE", "RUNNING", "ABORTED", "UNKNOWN"}

	for _, status := range statuses {
		result := StatusBadge(status)
		if result == "" {
			t.Errorf("StatusBadge(%s) returned empty string", status)
		}
	}
}

func TestStylesAreDefined(t *testing.T) {
	// Test that all exported styles are properly initialized
	styles := []lipgloss.Style{
		BaseStyle,
		DimStyle,
		TitleStyle,
		SubtitleStyle,
		SectionTitleStyle,
		ErrorStyle,
		SuccessStyle,
		WarningStyle,
		InfoStyle,
		RunningStyle,
		MutedStyle,
		AccentStyle,
		BoxStyle,
		FocusedBoxStyle,
		PanelStyle,
		CardStyle,
		KPICardStyle,
		TabBarStyle,
		StatusBarStyle,
		StatusBarKeyStyle,
		StatusBarDescStyle,
		TableHeaderStyle,
		TableRowStyle,
		TableRowAltStyle,
		TableSelectedStyle,
		TableCellMutedStyle,
		InputStyle,
		InputFocusedStyle,
		InputLabelStyle,
		InputPlaceholderStyle,
		SearchBarStyle,
		ButtonStyle,
		ButtonFocusedStyle,
		ButtonPrimaryStyle,
		ButtonDangerStyle,
		HelpStyle,
		HelpKeyStyle,
		HelpDescStyle,
		HelpSeparatorStyle,
	}

	for i, style := range styles {
		// Each style should render without panicking
		result := style.Render("test")
		if result == "" {
			t.Errorf("style %d rendered empty string", i)
		}
	}
}

func TestColorsAreDefined(t *testing.T) {
	colors := []lipgloss.Color{
		Primary,
		Secondary,
		Accent,
		Highlight,
		Background,
		Surface,
		SurfaceAlt,
		Foreground,
		ForegroundDim,
		Border,
		BorderFocus,
		Success,
		Error,
		Warning,
		Info,
		Muted,
		Running,
		Pending,
		Disabled,
		BuildSuccess,
		BuildFailure,
		BuildUnstable,
		BuildAborted,
		BuildRunning,
		BuildNotBuilt,
		BuildDisabled,
	}

	for i, color := range colors {
		// Each color should be a valid color string
		if string(color) == "" {
			t.Errorf("color %d is empty", i)
		}
	}
}

func TestIconsAreDefined(t *testing.T) {
	icons := []string{
		IconSuccess, IconFailure, IconWarning, IconRunning,
		IconPending, IconAborted, IconNotBuilt, IconDisabled, IconUnknown,
		IconArrowRight, IconArrowLeft, IconArrowUp, IconArrowDown,
		IconExpand, IconCollapse, IconFolder, IconFolderOpen, IconFile,
		IconSearch, IconFilter, IconRefresh, IconClock, IconCalendar,
		IconUser, IconServer, IconBuild, IconQueue, IconLog, IconArtifact,
		IconBranch, IconCommit, IconLink, IconStar, IconStarEmpty,
	}

	for i, icon := range icons {
		if icon == "" {
			t.Errorf("icon %d is empty", i)
		}
	}
}
