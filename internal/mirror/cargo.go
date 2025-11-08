package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CargoMirror handles Rust cargo registry configuration
type CargoMirror struct {
	registryURL string
}

// NewCargoMirror creates a new Cargo mirror handler
func NewCargoMirror(registryURL string) *CargoMirror {
	return &CargoMirror{
		registryURL: registryURL,
	}
}

// getCargoConfigPath returns the path to cargo config.toml
func getCargoConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// ~/.cargo/config.toml
	cargoDir := filepath.Join(homeDir, ".cargo")
	if err := os.MkdirAll(cargoDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cargo directory: %w", err)
	}

	return filepath.Join(cargoDir, "config.toml"), nil
}

// Enable configures cargo to use the mirror registry
func (c *CargoMirror) Enable() error {
	cargoConfigPath, err := getCargoConfigPath()
	if err != nil {
		return err
	}

	// Read existing config if it exists
	var existingContent string
	if data, err := os.ReadFile(cargoConfigPath); err == nil {
		existingContent = string(data)
	}

	// Check if source section exists
	lines := strings.Split(existingContent, "\n")
	hasSourceSection := false
	hasCratesIOSection := false
	newLines := []string{}

	inCratesIOSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "[source.crates-io]" {
			hasSourceSection = true
			hasCratesIOSection = true
			inCratesIOSection = true
			newLines = append(newLines, line)
			continue
		}

		if strings.HasPrefix(trimmed, "[source.") && trimmed != "[source.crates-io]" {
			inCratesIOSection = false
			newLines = append(newLines, line)
			continue
		}

		if strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "[source.") {
			// Leaving source sections
			if inCratesIOSection && !strings.Contains(existingContent, "replace-with") {
				newLines = append(newLines, fmt.Sprintf("replace-with = 'ustc'"))
			}
			inCratesIOSection = false
			newLines = append(newLines, line)
			continue
		}

		if inCratesIOSection && strings.HasPrefix(trimmed, "replace-with") {
			// Replace existing replace-with
			newLines = append(newLines, "replace-with = 'ustc'")
			continue
		}

		if trimmed != "" || len(newLines) == 0 {
			newLines = append(newLines, line)
		}
	}

	// Add configuration if it doesn't exist
	if !hasSourceSection || !hasCratesIOSection {
		if len(newLines) > 0 && newLines[len(newLines)-1] != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, "[source.crates-io]")
		newLines = append(newLines, "replace-with = 'ustc'")
		newLines = append(newLines, "")
		newLines = append(newLines, "[source.ustc]")
		newLines = append(newLines, fmt.Sprintf("registry = \"%s\"", c.registryURL))
	} else if !strings.Contains(existingContent, "[source.ustc]") {
		newLines = append(newLines, "")
		newLines = append(newLines, "[source.ustc]")
		newLines = append(newLines, fmt.Sprintf("registry = \"%s\"", c.registryURL))
	}

	// Write back
	content := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(cargoConfigPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write cargo config: %w", err)
	}

	return nil
}

// Disable removes the mirror configuration
func (c *CargoMirror) Disable() error {
	cargoConfigPath, err := getCargoConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(cargoConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cargo config: %w", err)
	}

	// Remove crosh-related configuration
	lines := strings.Split(string(data), "\n")
	newLines := []string{}
	skipSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "[source.crates-io]" || trimmed == "[source.ustc]" {
			skipSection = true
			continue
		}

		if strings.HasPrefix(trimmed, "[") {
			skipSection = false
		}

		if !skipSection && trimmed != "" {
			newLines = append(newLines, line)
		}
	}

	// Write back or remove file if empty
	if len(newLines) > 0 {
		content := strings.Join(newLines, "\n") + "\n"
		if err := os.WriteFile(cargoConfigPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write cargo config: %w", err)
		}
	} else {
		os.Remove(cargoConfigPath)
	}

	return nil
}

// Status checks if the mirror is currently enabled
func (c *CargoMirror) Status() (bool, string, error) {
	cargoConfigPath, err := getCargoConfigPath()
	if err != nil {
		return false, "", err
	}

	data, err := os.ReadFile(cargoConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "default registry", nil
		}
		return false, "", fmt.Errorf("failed to read cargo config: %w", err)
	}

	content := string(data)
	if strings.Contains(content, "[source.ustc]") {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "registry") {
				parts := strings.SplitN(trimmed, "=", 2)
				if len(parts) == 2 {
					registry := strings.Trim(strings.TrimSpace(parts[1]), "\"")
					return true, registry, nil
				}
			}
		}
	}

	return false, "default registry", nil
}
