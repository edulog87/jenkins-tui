// Package logger provides logging functionality for the application.
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	LogFile = "/tmp/jenkins-tui.log"
)

var (
	instance *slog.Logger
	once     sync.Once
	logFile  *os.File
)

// Init initializes the logger
func Init(debug bool) error {
	var initErr error
	once.Do(func() {
		// Create or truncate log file
		var err error
		logFile, err = os.OpenFile(LogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}

		level := slog.LevelInfo
		if debug {
			level = slog.LevelDebug
		}

		opts := &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Format time more readably
				if a.Key == slog.TimeKey {
					if t, ok := a.Value.Any().(time.Time); ok {
						a.Value = slog.StringValue(t.Format("15:04:05.000"))
					}
				}
				return a
			},
		}

		handler := slog.NewTextHandler(logFile, opts)
		instance = slog.New(handler)

		instance.Info("Logger initialized",
			"debug", debug,
			"logFile", LogFile,
		)
	})
	return initErr
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Sync()
		logFile.Close()
	}
}

// Get returns the logger instance
func Get() *slog.Logger {
	if instance == nil {
		// Fallback to stderr if not initialized
		return slog.Default()
	}
	return instance
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// LogFile returns the path to the log file
func LogFilePath() string {
	absPath, err := filepath.Abs(LogFile)
	if err != nil {
		return LogFile
	}
	return absPath
}
