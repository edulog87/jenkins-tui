package components

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/elogrono/jenkins-tui/internal/ui/theme"
)

func TestNewPanel(t *testing.T) {
	p := NewPanel("Test Panel", 50, 20)

	if p.Title != "Test Panel" {
		t.Errorf("Expected title 'Test Panel', got '%s'", p.Title)
	}
	if p.Width != 50 {
		t.Errorf("Expected width 50, got %d", p.Width)
	}
	if p.Height != 20 {
		t.Errorf("Expected height 20, got %d", p.Height)
	}
	if p.Focused {
		t.Error("Expected Focused to be false by default")
	}
}

func TestPanelSetContent(t *testing.T) {
	p := NewPanel("Test", 50, 20)
	result := p.SetContent("Hello World")

	if p.Content != "Hello World" {
		t.Errorf("Expected content 'Hello World', got '%s'", p.Content)
	}
	if result != p {
		t.Error("SetContent should return the panel for chaining")
	}
}

func TestPanelSetFocused(t *testing.T) {
	p := NewPanel("Test", 50, 20)
	result := p.SetFocused(true)

	if !p.Focused {
		t.Error("Expected Focused to be true")
	}
	if result != p {
		t.Error("SetFocused should return the panel for chaining")
	}
}

func TestPanelRender(t *testing.T) {
	p := NewPanel("Test Panel", 50, 10)
	p.SetContent("Content here")

	output := p.Render()

	// The output should contain the title
	if !strings.Contains(output, "Test Panel") {
		t.Error("Panel render should contain the title")
	}
	// The output should contain the content
	if !strings.Contains(output, "Content here") {
		t.Error("Panel render should contain the content")
	}
}

func TestNewKPICard(t *testing.T) {
	card := NewKPICard("Builds", "42", theme.IconBuild, theme.Primary)

	if card.Label != "Builds" {
		t.Errorf("Expected label 'Builds', got '%s'", card.Label)
	}
	if card.Value != "42" {
		t.Errorf("Expected value '42', got '%s'", card.Value)
	}
	if card.Icon != theme.IconBuild {
		t.Errorf("Expected icon '%s', got '%s'", theme.IconBuild, card.Icon)
	}
	if card.Width != 18 {
		t.Errorf("Expected default width 18, got %d", card.Width)
	}
}

func TestKPICardSetWidth(t *testing.T) {
	card := NewKPICard("Test", "0", "", theme.Primary)
	result := card.SetWidth(25)

	if card.Width != 25 {
		t.Errorf("Expected width 25, got %d", card.Width)
	}
	if result != card {
		t.Error("SetWidth should return the card for chaining")
	}
}

func TestKPICardRender(t *testing.T) {
	card := NewKPICard("Running", "5", theme.IconRunning, theme.Primary)
	output := card.Render()

	// Should contain the value and label
	if !strings.Contains(output, "5") {
		t.Error("KPICard render should contain the value")
	}
	if !strings.Contains(output, "Running") {
		t.Error("KPICard render should contain the label")
	}
}

func TestNewShortkeyBar(t *testing.T) {
	bar := NewShortkeyBar(80)

	if bar.Width != 80 {
		t.Errorf("Expected width 80, got %d", bar.Width)
	}
	if bar.Separator != "│" {
		t.Errorf("Expected separator '│', got '%s'", bar.Separator)
	}
	if len(bar.Keys) != 0 {
		t.Error("Expected empty keys slice")
	}
}

func TestShortkeyBarAdd(t *testing.T) {
	bar := NewShortkeyBar(80)
	result := bar.Add("Enter", "Select").Add("Esc", "Back")

	if len(bar.Keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(bar.Keys))
	}
	if bar.Keys[0].Key != "Enter" {
		t.Errorf("Expected first key 'Enter', got '%s'", bar.Keys[0].Key)
	}
	if bar.Keys[1].Description != "Back" {
		t.Errorf("Expected second description 'Back', got '%s'", bar.Keys[1].Description)
	}
	if result != bar {
		t.Error("Add should return the bar for chaining")
	}
}

func TestShortkeyBarRender(t *testing.T) {
	bar := NewShortkeyBar(80)
	bar.Add("Enter", "Select").Add("Esc", "Back")

	output := bar.Render()

	if !strings.Contains(output, "Enter") {
		t.Error("ShortkeyBar render should contain 'Enter'")
	}
	if !strings.Contains(output, "Select") {
		t.Error("ShortkeyBar render should contain 'Select'")
	}
	if !strings.Contains(output, "Esc") {
		t.Error("ShortkeyBar render should contain 'Esc'")
	}
}

func TestNewBreadcrumb(t *testing.T) {
	b := NewBreadcrumb("Views", "MyView", "Job")

	if len(b.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(b.Items))
	}
	if b.Items[0] != "Views" {
		t.Errorf("Expected first item 'Views', got '%s'", b.Items[0])
	}
}

func TestBreadcrumbRenderEmpty(t *testing.T) {
	b := NewBreadcrumb()
	output := b.Render()

	if output != "" {
		t.Errorf("Empty breadcrumb should render empty string, got '%s'", output)
	}
}

func TestBreadcrumbRender(t *testing.T) {
	b := NewBreadcrumb("Views", "MyView", "Job")
	output := b.Render()

	if !strings.Contains(output, "Views") {
		t.Error("Breadcrumb render should contain 'Views'")
	}
	if !strings.Contains(output, "MyView") {
		t.Error("Breadcrumb render should contain 'MyView'")
	}
	if !strings.Contains(output, "Job") {
		t.Error("Breadcrumb render should contain 'Job'")
	}
}

func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar(50, 20)

	if pb.Percent != 50 {
		t.Errorf("Expected percent 50, got %d", pb.Percent)
	}
	if pb.Width != 20 {
		t.Errorf("Expected width 20, got %d", pb.Width)
	}
	if !pb.ShowLabel {
		t.Error("Expected ShowLabel to be true by default")
	}
}

func TestProgressBarSetColor(t *testing.T) {
	pb := NewProgressBar(50, 20)
	result := pb.SetColor(lipgloss.Color("#FF0000"))

	if pb.Color != lipgloss.Color("#FF0000") {
		t.Error("SetColor did not set the color correctly")
	}
	if result != pb {
		t.Error("SetColor should return the progress bar for chaining")
	}
}

func TestProgressBarRender(t *testing.T) {
	pb := NewProgressBar(50, 10)
	output := pb.Render()

	// Should contain filled blocks and percentage
	if !strings.Contains(output, "█") {
		t.Error("Progress bar should contain filled blocks")
	}
	if !strings.Contains(output, "░") {
		t.Error("Progress bar should contain empty blocks")
	}
	if !strings.Contains(output, "50%") {
		t.Error("Progress bar should show 50%")
	}
}

func TestProgressBarRenderFull(t *testing.T) {
	pb := NewProgressBar(100, 10)
	output := pb.Render()

	if !strings.Contains(output, "100%") {
		t.Error("Full progress bar should show 100%")
	}
}

func TestProgressBarRenderOverflow(t *testing.T) {
	pb := NewProgressBar(150, 10)
	output := pb.Render()

	// Should cap at 100% visually
	filledCount := strings.Count(output, "█")
	emptyCount := strings.Count(output, "░")

	// With overflow, all blocks should be filled
	if emptyCount != 0 && filledCount != 10 {
		t.Errorf("Expected all 10 blocks filled for 150%%, got %d filled and %d empty", filledCount, emptyCount)
	}
}

func TestNewStagePipeline(t *testing.T) {
	sp := NewStagePipeline(100)

	if sp.Width != 100 {
		t.Errorf("Expected width 100, got %d", sp.Width)
	}
	if len(sp.Stages) != 0 {
		t.Error("Expected empty stages slice")
	}
}

func TestStagePipelineAddStage(t *testing.T) {
	sp := NewStagePipeline(100)
	result := sp.AddStage("Build", StageStatusSuccess, 5*time.Second)

	if len(sp.Stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(sp.Stages))
	}
	if sp.Stages[0].Name != "Build" {
		t.Errorf("Expected stage name 'Build', got '%s'", sp.Stages[0].Name)
	}
	if sp.Stages[0].Status != StageStatusSuccess {
		t.Errorf("Expected status SUCCESS, got '%s'", sp.Stages[0].Status)
	}
	if result != sp {
		t.Error("AddStage should return the pipeline for chaining")
	}
}

func TestStagePipelineRenderEmpty(t *testing.T) {
	sp := NewStagePipeline(100)
	output := sp.Render()

	if !strings.Contains(output, "No stages") {
		t.Error("Empty pipeline should render 'No stages'")
	}
}

func TestStagePipelineRender(t *testing.T) {
	sp := NewStagePipeline(100)
	sp.AddStage("Build", StageStatusSuccess, 10*time.Second)
	sp.AddStage("Test", StageStatusRunning, 5*time.Second)

	output := sp.Render()

	if !strings.Contains(output, "Build") {
		t.Error("Pipeline render should contain 'Build'")
	}
	if !strings.Contains(output, "Test") {
		t.Error("Pipeline render should contain 'Test'")
	}
}

func TestStatusIndicator(t *testing.T) {
	tests := []struct {
		status   string
		contains string
	}{
		{"SUCCESS", "SUCCESS"},
		{"FAILURE", "FAILURE"},
		{"UNSTABLE", "UNSTABLE"},
		{"RUNNING", "RUNNING"},
		{"ABORTED", "ABORTED"},
		{"PENDING", "PENDING"},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		output := StatusIndicator(tt.status)
		if !strings.Contains(output, tt.contains) {
			t.Errorf("StatusIndicator(%s) should contain '%s', got '%s'", tt.status, tt.contains, output)
		}
	}
}

func TestInfoRow(t *testing.T) {
	output := InfoRow("Status", "Running")

	if !strings.Contains(output, "Status") {
		t.Error("InfoRow should contain label")
	}
	if !strings.Contains(output, "Running") {
		t.Error("InfoRow should contain value")
	}
}

func TestInfoRowWithIcon(t *testing.T) {
	output := InfoRowWithIcon("►", "Status", "Running")

	if !strings.Contains(output, "Status") {
		t.Error("InfoRowWithIcon should contain label")
	}
	if !strings.Contains(output, "Running") {
		t.Error("InfoRowWithIcon should contain value")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "< 1s"},
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{3661 * time.Second, "1h 1m"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = '%s', want '%s'", tt.duration, result, tt.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"Hello", 10, "Hello"},
		{"Hello World", 8, "Hello..."},
		{"Hi", 5, "Hi"},
		{"Test", 3, "Tes"},
		{"AB", 1, "A"},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.max)
		if result != tt.expected {
			t.Errorf("truncate('%s', %d) = '%s', want '%s'", tt.input, tt.max, result, tt.expected)
		}
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		time     time.Time
		contains string
	}{
		{time.Time{}, "-"},
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5m ago"},
		{now.Add(-2 * time.Hour), "2h ago"},
		{now.Add(-3 * 24 * time.Hour), "3d ago"},
		{now.Add(-30 * 24 * time.Hour), ""},
	}

	for i, tt := range tests {
		result := FormatTimeAgo(tt.time)
		if tt.contains != "" && !strings.Contains(result, tt.contains) {
			t.Errorf("Test %d: FormatTimeAgo should contain '%s', got '%s'", i, tt.contains, result)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = '%s', want '%s'", tt.bytes, result, tt.expected)
		}
	}
}
