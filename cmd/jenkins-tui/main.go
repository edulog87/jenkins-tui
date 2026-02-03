package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/elogrono/jenkins-tui/internal/app"
	"github.com/elogrono/jenkins-tui/internal/config"
	"github.com/elogrono/jenkins-tui/internal/logger"
)

func main() {
	// Initialize logger
	if err := logger.Init(false); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create and run the program
	m := app.NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
