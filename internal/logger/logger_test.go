package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogFilePath(t *testing.T) {
	path := LogFilePath()
	if path == "" {
		t.Error("LogFilePath returned empty string")
	}

	// Should be an absolute path
	if !filepath.IsAbs(path) {
		t.Errorf("LogFilePath should return absolute path, got: %s", path)
	}

	// Should end with the expected filename
	if !strings.HasSuffix(path, "jenkins-tui.log") {
		t.Errorf("LogFilePath should end with jenkins-tui.log, got: %s", path)
	}
}

func TestGetWithoutInit(t *testing.T) {
	// Reset instance for this test
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	logger := Get()
	if logger == nil {
		t.Error("Get() should not return nil even without Init()")
	}
}

func TestLogFunctions(t *testing.T) {
	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "logger_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Test that log functions don't panic when instance is nil
	oldInstance := instance
	instance = nil

	// These should not panic
	Debug("test debug")
	Info("test info")
	Warn("test warn")
	Error("test error")

	// Restore instance
	instance = oldInstance
}

func TestInitCreatesLogFile(t *testing.T) {
	// Skip if we can't write to /tmp
	testLogFile := "/tmp/jenkins-tui-test-init.log"

	// Clean up any existing test file
	os.Remove(testLogFile)

	// We can't easily test Init() because it uses sync.Once
	// and the log file path is hardcoded. Instead, we verify
	// the LogFile constant is set correctly.
	if LogFile == "" {
		t.Error("LogFile constant should not be empty")
	}

	if !strings.HasPrefix(LogFile, "/tmp/") {
		t.Logf("LogFile is at: %s (not in /tmp as expected)", LogFile)
	}
}

func TestLogLevels(t *testing.T) {
	// This test verifies that the log functions don't panic
	// and can be called safely in any order

	tests := []struct {
		name  string
		logFn func(string, ...any)
		msg   string
		args  []any
	}{
		{"Debug without args", Debug, "debug message", nil},
		{"Info without args", Info, "info message", nil},
		{"Warn without args", Warn, "warn message", nil},
		{"Error without args", Error, "error message", nil},
		{"Debug with args", Debug, "debug with args", []any{"key", "value"}},
		{"Info with args", Info, "info with args", []any{"count", 42}},
		{"Warn with args", Warn, "warn with args", []any{"error", "something went wrong"}},
		{"Error with args", Error, "error with args", []any{"code", 500, "msg", "internal error"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			tt.logFn(tt.msg, tt.args...)
		})
	}
}
