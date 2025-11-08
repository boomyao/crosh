package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NPMMirror handles npm registry configuration
type NPMMirror struct {
	registryURL string
}

// NewNPMMirror creates a new NPM mirror handler
func NewNPMMirror(registryURL string) *NPMMirror {
	return &NPMMirror{
		registryURL: registryURL,
	}
}

// Enable configures npm to use the mirror registry
func (n *NPMMirror) Enable() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	npmrcPath := filepath.Join(homeDir, ".npmrc")

	// Read existing .npmrc file if it exists
	var existingContent string
	if data, err := os.ReadFile(npmrcPath); err == nil {
		existingContent = string(data)
	}

	// Check if registry is already configured
	lines := strings.Split(existingContent, "\n")
	registryLine := fmt.Sprintf("registry=%s", n.registryURL)
	hasRegistry := false
	newLines := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "registry=") {
			// Replace existing registry
			newLines = append(newLines, registryLine)
			hasRegistry = true
		} else if trimmed != "" {
			newLines = append(newLines, line)
		}
	}

	if !hasRegistry {
		newLines = append(newLines, registryLine)
	}

	// Write back to .npmrc
	content := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(npmrcPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .npmrc: %w", err)
	}

	return nil
}

// Disable removes the mirror configuration
func (n *NPMMirror) Disable() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	npmrcPath := filepath.Join(homeDir, ".npmrc")

	// Read existing .npmrc file
	data, err := os.ReadFile(npmrcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to disable
		}
		return fmt.Errorf("failed to read .npmrc: %w", err)
	}

	// Remove registry line
	lines := strings.Split(string(data), "\n")
	newLines := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "registry=") && trimmed != "" {
			newLines = append(newLines, line)
		}
	}

	// Write back
	if len(newLines) > 0 {
		content := strings.Join(newLines, "\n") + "\n"
		if err := os.WriteFile(npmrcPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write .npmrc: %w", err)
		}
	} else {
		// Remove file if empty
		os.Remove(npmrcPath)
	}

	return nil
}

// Status checks if the mirror is currently enabled
func (n *NPMMirror) Status() (bool, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	npmrcPath := filepath.Join(homeDir, ".npmrc")

	data, err := os.ReadFile(npmrcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "default registry", nil
		}
		return false, "", fmt.Errorf("failed to read .npmrc: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "registry=") {
			registry := strings.TrimPrefix(trimmed, "registry=")
			return true, registry, nil
		}
	}

	return false, "default registry", nil
}
