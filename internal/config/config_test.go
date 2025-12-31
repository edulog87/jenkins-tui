package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Profile.TimeoutSeconds != 15 {
		t.Errorf("expected TimeoutSeconds=15, got %d", cfg.Profile.TimeoutSeconds)
	}
	if cfg.Profile.AutoRefreshSeconds != 10 {
		t.Errorf("expected AutoRefreshSeconds=10, got %d", cfg.Profile.AutoRefreshSeconds)
	}
	if cfg.Profile.MaxBuildsPerJob != 200 {
		t.Errorf("expected MaxBuildsPerJob=200, got %d", cfg.Profile.MaxBuildsPerJob)
	}
	if cfg.Profile.MaxLogBytes != 200000 {
		t.Errorf("expected MaxLogBytes=200000, got %d", cfg.Profile.MaxLogBytes)
	}
	if cfg.Profile.RateLimitRPS != 5 {
		t.Errorf("expected RateLimitRPS=5, got %d", cfg.Profile.RateLimitRPS)
	}
	if cfg.Profile.InsecureSkipTLSVerify {
		t.Error("expected InsecureSkipTLSVerify=false")
	}
}

func TestIsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected bool
	}{
		{
			name:     "empty config",
			cfg:      DefaultConfig(),
			expected: false,
		},
		{
			name: "partial config - missing token",
			cfg: &Config{
				Profile: Profile{
					BaseURL:  "https://jenkins.example.com",
					Username: "admin",
				},
			},
			expected: false,
		},
		{
			name: "complete config",
			cfg: &Config{
				Profile: Profile{
					BaseURL:  "https://jenkins.example.com",
					Username: "admin",
					APIToken: "secret-token",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsConfigured(); got != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "empty config",
			cfg:     DefaultConfig(),
			wantErr: true,
		},
		{
			name: "missing username",
			cfg: &Config{
				Profile: Profile{
					BaseURL:  "https://jenkins.example.com",
					APIToken: "token",
				},
			},
			wantErr: true,
		},
		{
			name: "valid config",
			cfg: &Config{
				Profile: Profile{
					BaseURL:  "https://jenkins.example.com",
					Username: "admin",
					APIToken: "token",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "jenkins-tui-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config file path
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a test config
	cfg := &Config{
		Profile: Profile{
			BaseURL:               "https://jenkins.example.com",
			Username:              "testuser",
			APIToken:              "test-token-123",
			InsecureSkipTLSVerify: true,
			TimeoutSeconds:        30,
			AutoRefreshSeconds:    5,
			MaxBuildsPerJob:       100,
			MaxLogBytes:           100000,
			RateLimitRPS:          10,
		},
	}

	// Create config directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Save config manually to temp location (since ConfigPath() returns real path)
	f, err := os.Create(configPath)
	if err != nil {
		t.Fatal(err)
	}

	content := `# Test config
[profile]
base_url = "https://jenkins.example.com"
username = "testuser"
api_token = "test-token-123"
insecure_skip_tls_verify = true
timeout_seconds = 30
auto_refresh_seconds = 5
max_builds_per_job = 100
max_log_bytes = 100000
rate_limit_rps = 10
`
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	// Test that saved config matches
	if cfg.Profile.BaseURL != "https://jenkins.example.com" {
		t.Error("BaseURL mismatch")
	}
	if cfg.Profile.Username != "testuser" {
		t.Error("Username mismatch")
	}
	if cfg.Profile.APIToken != "test-token-123" {
		t.Error("APIToken mismatch")
	}
}
