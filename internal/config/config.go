// Package config handles application configuration loading, validation and persistence.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	Profile Profile `toml:"profile"`
}

// Profile represents a Jenkins server connection profile
type Profile struct {
	BaseURL               string `toml:"base_url"`
	Username              string `toml:"username"`
	APIToken              string `toml:"api_token"`
	InsecureSkipTLSVerify bool   `toml:"insecure_skip_tls_verify"`
	TimeoutSeconds        int    `toml:"timeout_seconds"`
	AutoRefreshSeconds    int    `toml:"auto_refresh_seconds"`
	MaxBuildsPerJob       int    `toml:"max_builds_per_job"`
	MaxLogBytes           int    `toml:"max_log_bytes"`
	RateLimitRPS          int    `toml:"rate_limit_rps"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Profile: Profile{
			InsecureSkipTLSVerify: false,
			TimeoutSeconds:        15,
			AutoRefreshSeconds:    10,
			MaxBuildsPerJob:       200,
			MaxLogBytes:           200000,
			RateLimitRPS:          5,
		},
	}
}

// ConfigPath returns the path to the configuration file
func ConfigPath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
	default: // linux, darwin
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to determine home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configDir, "jenkins-tui", "config.toml"), nil
}

// Load loads the configuration from the config file
// If the file doesn't exist, returns a default config
func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config - the app will prompt for setup
		return DefaultConfig(), nil
	}

	// Load existing config
	cfg := DefaultConfig()
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Create/overwrite config file
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}
	defer f.Close()

	// Write header comment
	if _, err := f.WriteString("# Jenkins TUI Configuration\n\n"); err != nil {
		return err
	}

	// Encode config as TOML
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	return nil
}

// IsConfigured returns true if the essential fields are set
func (c *Config) IsConfigured() bool {
	return c.Profile.BaseURL != "" &&
		c.Profile.Username != "" &&
		c.Profile.APIToken != ""
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Profile.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if c.Profile.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.Profile.APIToken == "" {
		return fmt.Errorf("api_token is required")
	}
	if c.Profile.TimeoutSeconds <= 0 {
		c.Profile.TimeoutSeconds = 15
	}
	if c.Profile.AutoRefreshSeconds <= 0 {
		c.Profile.AutoRefreshSeconds = 10
	}
	if c.Profile.MaxBuildsPerJob <= 0 {
		c.Profile.MaxBuildsPerJob = 200
	}
	if c.Profile.MaxLogBytes <= 0 {
		c.Profile.MaxLogBytes = 200000
	}
	if c.Profile.RateLimitRPS <= 0 {
		c.Profile.RateLimitRPS = 5
	}
	return nil
}
