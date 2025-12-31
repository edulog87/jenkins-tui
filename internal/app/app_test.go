package app

import (
	"testing"
	"time"

	"github.com/elogrono/jenkins-tui/internal/config"
)

func TestNewModel(t *testing.T) {
	cfg := config.DefaultConfig()
	model := NewModel(cfg)

	if model == nil {
		t.Fatal("NewModel returned nil")
	}

	// Should be in setup state since config is not configured
	if model.state != StateSetup {
		t.Errorf("expected state StateSetup, got %v", model.state)
	}

	if model.setupModel == nil {
		t.Error("expected setupModel to be initialized for unconfigured config")
	}
}

func TestNewModelWithConfig(t *testing.T) {
	cfg := &config.Config{
		Profile: config.Profile{
			BaseURL:            "https://jenkins.example.com",
			Username:           "admin",
			APIToken:           "token123",
			TimeoutSeconds:     15,
			AutoRefreshSeconds: 10,
			MaxBuildsPerJob:    200,
			MaxLogBytes:        200000,
			RateLimitRPS:       5,
		},
	}
	model := NewModel(cfg)

	if model == nil {
		t.Fatal("NewModel returned nil")
	}

	// Should be in loading state since config is configured
	if model.state != StateLoading {
		t.Errorf("expected state StateLoading, got %v", model.state)
	}

	if model.setupModel != nil {
		t.Error("expected setupModel to be nil for configured config")
	}
}

func TestNewSetupModel(t *testing.T) {
	setup := NewSetupModel()

	if setup == nil {
		t.Fatal("NewSetupModel returned nil")
	}

	if setup.focusedField != FieldURL {
		t.Errorf("expected focusedField to be FieldURL, got %v", setup.focusedField)
	}
}

func TestSetupModelNextField(t *testing.T) {
	setup := NewSetupModel()

	// Start at URL
	if setup.focusedField != FieldURL {
		t.Errorf("expected FieldURL, got %v", setup.focusedField)
	}

	// Move to username
	setup.nextField()
	if setup.focusedField != FieldUsername {
		t.Errorf("expected FieldUsername, got %v", setup.focusedField)
	}

	// Move to token
	setup.nextField()
	if setup.focusedField != FieldToken {
		t.Errorf("expected FieldToken, got %v", setup.focusedField)
	}

	// Move to submit
	setup.nextField()
	if setup.focusedField != FieldSubmit {
		t.Errorf("expected FieldSubmit, got %v", setup.focusedField)
	}

	// Wrap around to URL
	setup.nextField()
	if setup.focusedField != FieldURL {
		t.Errorf("expected FieldURL after wrap, got %v", setup.focusedField)
	}
}

func TestSetupModelPrevField(t *testing.T) {
	setup := NewSetupModel()

	// Start at URL, go back wraps to submit
	setup.prevField()
	if setup.focusedField != FieldSubmit {
		t.Errorf("expected FieldSubmit, got %v", setup.focusedField)
	}

	// Go back to token
	setup.prevField()
	if setup.focusedField != FieldToken {
		t.Errorf("expected FieldToken, got %v", setup.focusedField)
	}
}

func TestTabID(t *testing.T) {
	// Verify tab constants
	if TabDashboard != 0 {
		t.Errorf("expected TabDashboard=0, got %d", TabDashboard)
	}
	if TabViews != 1 {
		t.Errorf("expected TabViews=1, got %d", TabViews)
	}
	if TabBuilds != 2 {
		t.Errorf("expected TabBuilds=2, got %d", TabBuilds)
	}
}

func TestAppState(t *testing.T) {
	// Verify state constants
	if StateSetup != 0 {
		t.Errorf("expected StateSetup=0, got %d", StateSetup)
	}
	if StateLoading != 1 {
		t.Errorf("expected StateLoading=1, got %d", StateLoading)
	}
	if StateReady != 2 {
		t.Errorf("expected StateReady=2, got %d", StateReady)
	}
	if StateError != 3 {
		t.Errorf("expected StateError=3, got %d", StateError)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test truncate
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exact", 5, "exact"},
		{"", 10, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.max)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, result, tt.expected)
		}
	}

	// Test min
	if min(5, 10) != 5 {
		t.Error("min(5, 10) should be 5")
	}
	if min(10, 5) != 5 {
		t.Error("min(10, 5) should be 5")
	}
	if min(5, 5) != 5 {
		t.Error("min(5, 5) should be 5")
	}
}

func TestMinInt(t *testing.T) {
	if minInt(3, 7) != 3 {
		t.Error("minInt(3, 7) should be 3")
	}
	if minInt(7, 3) != 3 {
		t.Error("minInt(7, 3) should be 3")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{500 * time.Millisecond, "< 1s"},
		{1 * time.Second, "1s"},
		{30 * time.Second, "30s"},
		{1 * time.Minute, "1m 0s"},
		{90 * time.Second, "1m 30s"},
		{1 * time.Hour, "1h 0m"},
		{61 * time.Minute, "1h 1m"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.d)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, result, tt.expected)
		}
	}
}

func TestStatusToColor(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"SUCCESS", "blue"},
		{"FAILURE", "red"},
		{"UNSTABLE", "yellow"},
		{"ABORTED", "aborted"},
		{"RUNNING", "blue_anime"},
		{"UNKNOWN", "notbuilt"},
	}

	for _, tt := range tests {
		result := statusToColor(tt.status)
		if result != tt.expected {
			t.Errorf("statusToColor(%q) = %q, want %q", tt.status, result, tt.expected)
		}
	}
}
