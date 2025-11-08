package mirror

import (
	"fmt"
	"os"
	"strings"
)

// GoMirror handles Go module proxy configuration
type GoMirror struct {
	proxyURL string
}

// NewGoMirror creates a new Go mirror handler
func NewGoMirror(proxyURL string) *GoMirror {
	return &GoMirror{
		proxyURL: proxyURL,
	}
}

// Enable configures Go to use the mirror proxy
// This is done via environment variable GOPROXY
func (g *GoMirror) Enable() error {
	// For Go, we typically set environment variables
	// This will output the command to set the environment variable
	fmt.Printf("# Run the following command to enable Go proxy:\n")
	fmt.Printf("export GOPROXY=%s\n", g.proxyURL)
	fmt.Printf("# To make it permanent, add it to your ~/.bashrc or ~/.zshrc\n")

	// We can also try to append to shell rc files
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Try to detect shell
	shell := os.Getenv("SHELL")
	var rcFile string

	if strings.Contains(shell, "zsh") {
		rcFile = fmt.Sprintf("%s/.zshrc", homeDir)
	} else if strings.Contains(shell, "bash") {
		rcFile = fmt.Sprintf("%s/.bashrc", homeDir)
	} else {
		// Default to bashrc
		rcFile = fmt.Sprintf("%s/.bashrc", homeDir)
	}

	// Read existing rc file
	var existingContent string
	if data, err := os.ReadFile(rcFile); err == nil {
		existingContent = string(data)
	}

	// Check if GOPROXY is already set
	exportLine := fmt.Sprintf("export GOPROXY=%s", g.proxyURL)
	if strings.Contains(existingContent, "export GOPROXY=") {
		// Replace existing GOPROXY
		lines := strings.Split(existingContent, "\n")
		newLines := []string{}
		for _, line := range lines {
			if strings.Contains(line, "export GOPROXY=") {
				newLines = append(newLines, exportLine)
			} else {
				newLines = append(newLines, line)
			}
		}
		existingContent = strings.Join(newLines, "\n")
	} else {
		// Append new GOPROXY
		if !strings.HasSuffix(existingContent, "\n") {
			existingContent += "\n"
		}
		existingContent += fmt.Sprintf("\n# Added by crosh\n%s\n", exportLine)
	}

	// Write back
	if err := os.WriteFile(rcFile, []byte(existingContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", rcFile, err)
	}

	// Set for current session
	os.Setenv("GOPROXY", g.proxyURL)

	return nil
}

// Disable removes the Go proxy configuration
func (g *GoMirror) Disable() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	shell := os.Getenv("SHELL")
	var rcFile string

	if strings.Contains(shell, "zsh") {
		rcFile = fmt.Sprintf("%s/.zshrc", homeDir)
	} else if strings.Contains(shell, "bash") {
		rcFile = fmt.Sprintf("%s/.bashrc", homeDir)
	} else {
		rcFile = fmt.Sprintf("%s/.bashrc", homeDir)
	}

	data, err := os.ReadFile(rcFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", rcFile, err)
	}

	// Remove GOPROXY lines
	lines := strings.Split(string(data), "\n")
	newLines := []string{}
	skipNext := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "# Added by crosh" {
			skipNext = true
			continue
		}
		if skipNext && strings.Contains(line, "export GOPROXY=") {
			skipNext = false
			continue
		}
		if !strings.Contains(line, "export GOPROXY=") {
			newLines = append(newLines, line)
		}
	}

	// Write back
	content := strings.Join(newLines, "\n")
	if err := os.WriteFile(rcFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", rcFile, err)
	}

	// Unset for current session
	os.Unsetenv("GOPROXY")

	return nil
}

// Status checks if the Go proxy is currently enabled
func (g *GoMirror) Status() (bool, string, error) {
	goproxy := os.Getenv("GOPROXY")
	if goproxy != "" {
		return true, goproxy, nil
	}

	return false, "default proxy", nil
}

// GetEnvCommand returns the command to set environment variable for current session
func (g *GoMirror) GetEnvCommand() string {
	return fmt.Sprintf("export GOPROXY=%s", g.proxyURL)
}
