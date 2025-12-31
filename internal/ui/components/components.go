// Package components provides reusable UI components for the TUI.
package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/elogrono/jenkins-tui/internal/ui/theme"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PANEL COMPONENT
// ═══════════════════════════════════════════════════════════════════════════════

// Panel represents a titled panel with content
type Panel struct {
	Title   string
	Content string
	Width   int
	Height  int
	Focused bool
	Style   lipgloss.Style
}

// NewPanel creates a new panel
func NewPanel(title string, width, height int) *Panel {
	return &Panel{
		Title:  title,
		Width:  width,
		Height: height,
		Style:  theme.PanelStyle,
	}
}

// SetContent sets the panel content
func (p *Panel) SetContent(content string) *Panel {
	p.Content = content
	return p
}

// SetFocused sets the focus state
func (p *Panel) SetFocused(focused bool) *Panel {
	p.Focused = focused
	return p
}

// Render renders the panel
func (p *Panel) Render() string {
	style := p.Style
	if p.Focused {
		style = style.BorderForeground(theme.BorderFocus)
	}

	titleStyle := theme.SectionTitleStyle
	if p.Focused {
		titleStyle = titleStyle.Foreground(theme.Primary)
	}

	title := titleStyle.Render(p.Title)

	contentStyle := lipgloss.NewStyle().
		Width(p.Width - 4).
		Height(p.Height - 3)

	content := contentStyle.Render(p.Content)

	inner := lipgloss.JoinVertical(lipgloss.Left, title, content)

	return style.
		Width(p.Width).
		Height(p.Height).
		Render(inner)
}

// ═══════════════════════════════════════════════════════════════════════════════
// KPI CARD COMPONENT
// ═══════════════════════════════════════════════════════════════════════════════

// KPICard represents a KPI metric card
type KPICard struct {
	Label string
	Value string
	Icon  string
	Color lipgloss.Color
	Width int
}

// NewKPICard creates a new KPI card
func NewKPICard(label, value, icon string, color lipgloss.Color) *KPICard {
	return &KPICard{
		Label: label,
		Value: value,
		Icon:  icon,
		Color: color,
		Width: 18,
	}
}

// SetWidth sets the card width
func (k *KPICard) SetWidth(width int) *KPICard {
	k.Width = width
	return k
}

// Render renders the KPI card
func (k *KPICard) Render() string {
	valueStyle := lipgloss.NewStyle().
		Foreground(k.Color).
		Bold(true).
		Align(lipgloss.Center)

	labelStyle := lipgloss.NewStyle().
		Foreground(theme.ForegroundDim).
		Align(lipgloss.Center)

	iconStyle := lipgloss.NewStyle().
		Foreground(k.Color).
		Align(lipgloss.Center)

	content := lipgloss.JoinVertical(lipgloss.Center,
		iconStyle.Render(k.Icon),
		valueStyle.Render(k.Value),
		labelStyle.Render(k.Label),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(0, 1).
		Width(k.Width).
		Align(lipgloss.Center).
		Render(content)
}

// ═══════════════════════════════════════════════════════════════════════════════
// SHORTKEY BAR COMPONENT
// ═══════════════════════════════════════════════════════════════════════════════

// Shortkey represents a keyboard shortcut
type Shortkey struct {
	Key         string
	Description string
}

// ShortkeyBar represents a bar of shortcuts
type ShortkeyBar struct {
	Keys      []Shortkey
	Width     int
	Separator string
}

// NewShortkeyBar creates a new shortkey bar
func NewShortkeyBar(width int) *ShortkeyBar {
	return &ShortkeyBar{
		Width:     width,
		Separator: "│",
	}
}

// Add adds a shortkey to the bar
func (s *ShortkeyBar) Add(key, description string) *ShortkeyBar {
	s.Keys = append(s.Keys, Shortkey{Key: key, Description: description})
	return s
}

// Render renders the shortkey bar
func (s *ShortkeyBar) Render() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(theme.ForegroundDim)

	sepStyle := lipgloss.NewStyle().
		Foreground(theme.Border)

	var items []string
	for _, k := range s.Keys {
		items = append(items, keyStyle.Render(k.Key)+" "+descStyle.Render(k.Description))
	}

	bar := strings.Join(items, sepStyle.Render(" "+s.Separator+" "))

	return lipgloss.NewStyle().
		Background(theme.Surface).
		Width(s.Width).
		Padding(0, 1).
		Render(bar)
}

// ═══════════════════════════════════════════════════════════════════════════════
// BREADCRUMB COMPONENT
// ═══════════════════════════════════════════════════════════════════════════════

// Breadcrumb represents a navigation breadcrumb
type Breadcrumb struct {
	Items []string
}

// NewBreadcrumb creates a new breadcrumb
func NewBreadcrumb(items ...string) *Breadcrumb {
	return &Breadcrumb{Items: items}
}

// Render renders the breadcrumb
func (b *Breadcrumb) Render() string {
	if len(b.Items) == 0 {
		return ""
	}

	activeStyle := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(theme.ForegroundDim)

	sepStyle := lipgloss.NewStyle().
		Foreground(theme.Border)

	var parts []string
	for i, item := range b.Items {
		if i == len(b.Items)-1 {
			parts = append(parts, activeStyle.Render(item))
		} else {
			parts = append(parts, inactiveStyle.Render(item))
			parts = append(parts, sepStyle.Render(" "+theme.IconArrowRight+" "))
		}
	}

	return strings.Join(parts, "")
}

// ═══════════════════════════════════════════════════════════════════════════════
// PROGRESS BAR COMPONENT
// ═══════════════════════════════════════════════════════════════════════════════

// ProgressBar represents a progress bar
type ProgressBar struct {
	Percent   int
	Width     int
	ShowLabel bool
	Color     lipgloss.Color
}

// NewProgressBar creates a new progress bar
func NewProgressBar(percent, width int) *ProgressBar {
	return &ProgressBar{
		Percent:   percent,
		Width:     width,
		ShowLabel: true,
		Color:     theme.Primary,
	}
}

// SetColor sets the bar color
func (p *ProgressBar) SetColor(color lipgloss.Color) *ProgressBar {
	p.Color = color
	return p
}

// Render renders the progress bar
func (p *ProgressBar) Render() string {
	filled := (p.Percent * p.Width) / 100
	if filled > p.Width {
		filled = p.Width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.Width-filled)

	barStyle := lipgloss.NewStyle().Foreground(p.Color)

	if p.ShowLabel {
		label := fmt.Sprintf(" %3d%%", p.Percent)
		return barStyle.Render(bar) + theme.DimStyle.Render(label)
	}

	return barStyle.Render(bar)
}

// ═══════════════════════════════════════════════════════════════════════════════
// STAGE PIPELINE COMPONENT
// ═══════════════════════════════════════════════════════════════════════════════

// StageStatus represents the status of a pipeline stage
type StageStatus string

const (
	StageStatusSuccess  StageStatus = "SUCCESS"
	StageStatusFailure  StageStatus = "FAILURE"
	StageStatusRunning  StageStatus = "RUNNING"
	StageStatusPending  StageStatus = "PENDING"
	StageStatusSkipped  StageStatus = "SKIPPED"
	StageStatusUnstable StageStatus = "UNSTABLE"
	StageStatusAborted  StageStatus = "ABORTED"
)

// Stage represents a pipeline stage for display
type Stage struct {
	Name     string
	Status   StageStatus
	Duration time.Duration
}

// StagePipeline represents a pipeline visualization
type StagePipeline struct {
	Stages []Stage
	Width  int
}

// NewStagePipeline creates a new stage pipeline
func NewStagePipeline(width int) *StagePipeline {
	return &StagePipeline{
		Width: width,
	}
}

// AddStage adds a stage to the pipeline
func (s *StagePipeline) AddStage(name string, status StageStatus, duration time.Duration) *StagePipeline {
	s.Stages = append(s.Stages, Stage{
		Name:     name,
		Status:   status,
		Duration: duration,
	})
	return s
}

// Render renders the pipeline stages
func (s *StagePipeline) Render() string {
	if len(s.Stages) == 0 {
		return theme.MutedStyle.Render("No stages")
	}

	var stageBoxes []string

	for i, stage := range s.Stages {
		// Stage icon and color based on status
		var icon string
		var color lipgloss.Color

		switch stage.Status {
		case StageStatusSuccess:
			icon = theme.IconSuccess
			color = theme.BuildSuccess
		case StageStatusFailure:
			icon = theme.IconFailure
			color = theme.BuildFailure
		case StageStatusRunning:
			icon = theme.IconRunning
			color = theme.BuildRunning
		case StageStatusPending:
			icon = theme.IconPending
			color = theme.Muted
		case StageStatusSkipped:
			icon = theme.IconAborted
			color = theme.BuildAborted
		case StageStatusUnstable:
			icon = theme.IconWarning
			color = theme.BuildUnstable
		case StageStatusAborted:
			icon = theme.IconAborted
			color = theme.BuildAborted
		default:
			icon = theme.IconUnknown
			color = theme.Muted
		}

		iconStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
		nameStyle := lipgloss.NewStyle().Foreground(theme.Foreground)
		durStyle := lipgloss.NewStyle().Foreground(theme.ForegroundDim)

		// Format duration
		durStr := formatDuration(stage.Duration)

		stageContent := lipgloss.JoinVertical(lipgloss.Center,
			iconStyle.Render(icon),
			nameStyle.Render(truncate(stage.Name, 12)),
			durStyle.Render(durStr),
		)

		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(color).
			Padding(0, 1).
			Width(14).
			Align(lipgloss.Center)

		stageBoxes = append(stageBoxes, boxStyle.Render(stageContent))

		// Add connector except for last stage
		if i < len(s.Stages)-1 {
			connector := lipgloss.NewStyle().
				Foreground(theme.Border).
				Render("──▶")
			stageBoxes = append(stageBoxes, lipgloss.NewStyle().
				Padding(1, 0).
				Render(connector))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, stageBoxes...)
}

// ═══════════════════════════════════════════════════════════════════════════════
// STATUS INDICATOR COMPONENT
// ═══════════════════════════════════════════════════════════════════════════════

// StatusIndicator renders a status with icon
func StatusIndicator(status string) string {
	var icon string
	var style lipgloss.Style

	switch status {
	case "SUCCESS":
		icon = theme.IconSuccess
		style = theme.SuccessStyle
	case "FAILURE":
		icon = theme.IconFailure
		style = theme.ErrorStyle
	case "UNSTABLE":
		icon = theme.IconWarning
		style = theme.WarningStyle
	case "RUNNING", "BUILDING":
		icon = theme.IconRunning
		style = theme.RunningStyle
	case "ABORTED":
		icon = theme.IconAborted
		style = theme.MutedStyle
	case "PENDING", "NOT_BUILT":
		icon = theme.IconPending
		style = theme.MutedStyle
	default:
		icon = theme.IconUnknown
		style = theme.MutedStyle
	}

	return style.Render(icon + " " + status)
}

// ═══════════════════════════════════════════════════════════════════════════════
// INFO ROW COMPONENT
// ═══════════════════════════════════════════════════════════════════════════════

// InfoRow renders a label-value row
func InfoRow(label, value string) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(theme.ForegroundDim).
		Width(15)

	valueStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground)

	return labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

// InfoRowWithIcon renders a label-value row with icon
func InfoRowWithIcon(icon, label, value string) string {
	iconStyle := lipgloss.NewStyle().
		Foreground(theme.Primary)

	labelStyle := lipgloss.NewStyle().
		Foreground(theme.ForegroundDim).
		Width(15)

	valueStyle := lipgloss.NewStyle().
		Foreground(theme.Foreground)

	return iconStyle.Render(icon) + " " + labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

// ═══════════════════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════════

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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// FormatTimeAgo formats a time as relative (e.g., "5m ago", "2h ago")
func FormatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2")
	}
}

// FormatBytes formats bytes to human readable
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
