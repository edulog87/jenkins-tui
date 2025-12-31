// Package browser provides cross-platform browser opening functionality.
package browser

import (
	"os/exec"
	"runtime"
)

// Open opens the specified URL in the default browser.
// It works on Linux, macOS, and Windows.
func Open(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
