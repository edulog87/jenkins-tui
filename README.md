# Jenkins Go TUI

[![Go Version](https://img.shields.io/github/go-mod/go-version/elogrono/jenkins-tui)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A modern, fast, and reactive Terminal User Interface (TUI) for Jenkins, built with Go and the Bubble Tea framework. Monitor your builds, manage jobs, and inspect logs without leaving your terminal.

![Jenkins TUI Preview](https://via.placeholder.com/800x400?text=Jenkins+Go+TUI+Preview+Placeholder)

## 🚀 Features

- **Real-time Dashboard**: Monitor Jenkins health, running builds, build queue, and node status at a glance.
- **Job Exploration**: Browse through Jenkins views and jobs with incremental search/filtering.
- **Build History**: Paged history for jobs with detailed build results and duration.
- **Log Viewer**: Integrated log viewer with auto-scroll (tail/follow) and search capabilities.
- **Multi-Profile Support**: Manage multiple Jenkins instances with easy switching.
- **Safe Operations**: Read-only by default. Actions like rebuilding or aborting require explicit confirmation.
- **Keyboard-Driven**: Optimized for speed with intuitive keybindings.

## 🛠 Tech Stack

- **Core**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (MVU architecture)
- **UI Components**: [Bubbles](https://github.com/charmbracelet/bubbles)
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Jenkins Integration**: Official REST/JSON API with CSRF (Crumb) support.

## 📦 Installation

### From Source
Requires Go 1.22+
```bash
go install github.com/elogrono/jenkins-tui@latest
```

### Manual Build
```bash
git clone https://github.com/edulog87/jenkins-tui.git
cd jenkins-tui
# Note: if cmd/jenkins-tui/main.go is missing, it must be created as the entry point.
make build
```

## 📝 Extra Installation Steps
During the installation process, the following actions were performed to ensure the project works correctly:
1. Created `cmd/jenkins-tui/main.go` which was missing from the repository to serve as the application entry point.
2. Fixed `logger.Init` call and `config.Load` return values in `main.go` to match the internal implementation.
3. Installed the binary to `/home/elogrono/.local/bin` and updated the `PATH` in `.bashrc`.


## ⚙️ Configuration

The configuration file is automatically created on the first run.

- **Linux/macOS**: `~/.config/jenkins-tui/config.toml`
- **Windows**: `%APPDATA%\jenkins-tui\config.toml`

### Example `config.toml`

```toml
active_profile = "production"

[profiles.production]
base_url = "https://jenkins.example.com"
username = "your-user"
api_token = "your-api-token" # Recommended: Use env var JENKINS_TOKEN instead
auto_refresh_seconds = 10
timeout_seconds = 15
```

## ⌨️ Keybindings

### Navigation
- `Tab` / `Shift+Tab`: Switch between tabs.
- `1` - `9`: Jump directly to a tab.
- `Esc`: Go back or close modals.
- `?`: Toggle contextual help.

### General Actions
- `r`: Manual refresh.
- `/`: Activate search/filtering.
- `Enter`: Select item or view details.

### Logs
- `l`: Open logs for the selected build.
- `s`: Toggle "Follow" (tail) mode.
- `G` / `g`: Jump to bottom/top of logs.

## 🤝 Contributing

Contributions are welcome! Please check our [AGENTS.md](AGENTS.md) for architectural guidelines and development standards.

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/amazing-feature`).
3. Commit your changes (`git commit -m 'feat: add some amazing feature'`).
4. Push to the branch (`git push origin feature/amazing-feature`).
5. Open a Pull Request.

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
