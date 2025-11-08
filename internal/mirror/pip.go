package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PipMirror handles pip index configuration
type PipMirror struct {
	indexURL string
}

// NewPipMirror creates a new Pip mirror handler
func NewPipMirror(indexURL string) *PipMirror {
	return &PipMirror{
		indexURL: indexURL,
	}
}

// getPipConfigPath returns the path to pip.conf or pip.ini
func getPipConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Linux/macOS: ~/.config/pip/pip.conf
	configDir := filepath.Join(homeDir, ".config", "pip")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create pip config directory: %w", err)
	}

	return filepath.Join(configDir, "pip.conf"), nil
}

// Enable configures pip to use the mirror index
func (p *PipMirror) Enable() error {
	pipConfigPath, err := getPipConfigPath()
	if err != nil {
		return err
	}

	// Read existing config if it exists
	var existingContent string
	if data, err := os.ReadFile(pipConfigPath); err == nil {
		existingContent = string(data)
	}

	// Parse or create [global] section
	lines := strings.Split(existingContent, "\n")
	hasGlobalSection := false
	hasIndexURL := false
	newLines := []string{}
	inGlobalSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "[global]" {
			hasGlobalSection = true
			inGlobalSection = true
			newLines = append(newLines, line)
			continue
		}

		if strings.HasPrefix(trimmed, "[") && trimmed != "[global]" {
			// Entering a different section
			if inGlobalSection && !hasIndexURL {
				// Add index-url before leaving global section
				newLines = append(newLines, fmt.Sprintf("index-url = %s", p.indexURL))
				hasIndexURL = true
			}
			inGlobalSection = false
			newLines = append(newLines, line)
			continue
		}

		if inGlobalSection && strings.HasPrefix(trimmed, "index-url") {
			// Replace existing index-url
			newLines = append(newLines, fmt.Sprintf("index-url = %s", p.indexURL))
			hasIndexURL = true
			continue
		}

		if trimmed != "" {
			newLines = append(newLines, line)
		}
	}

	// Add [global] section if it doesn't exist
	if !hasGlobalSection {
		newLines = append(newLines, "[global]")
		newLines = append(newLines, fmt.Sprintf("index-url = %s", p.indexURL))
	} else if !hasIndexURL {
		// Add index-url to existing global section
		newLines = append(newLines, fmt.Sprintf("index-url = %s", p.indexURL))
	}

	// Write back
	content := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(pipConfigPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write pip config: %w", err)
	}

	return nil
}

// Disable removes the mirror configuration
func (p *PipMirror) Disable() error {
	pipConfigPath, err := getPipConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(pipConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read pip config: %w", err)
	}

	// Remove index-url line
	lines := strings.Split(string(data), "\n")
	newLines := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "index-url") && trimmed != "" {
			newLines = append(newLines, line)
		}
	}

	// Write back or remove file if empty
	if len(newLines) > 0 {
		content := strings.Join(newLines, "\n") + "\n"
		if err := os.WriteFile(pipConfigPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write pip config: %w", err)
		}
	} else {
		os.Remove(pipConfigPath)
	}

	return nil
}

// Status checks if the mirror is currently enabled
func (p *PipMirror) Status() (bool, string, error) {
	pipConfigPath, err := getPipConfigPath()
	if err != nil {
		return false, "", err
	}

	data, err := os.ReadFile(pipConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "default index", nil
		}
		return false, "", fmt.Errorf("failed to read pip config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "index-url") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				indexURL := strings.TrimSpace(parts[1])
				return true, indexURL, nil
			}
		}
	}

	return false, "default index", nil
}
