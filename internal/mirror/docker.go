package mirror

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DockerMirror handles Docker registry mirror configuration
type DockerMirror struct {
	registries []string
}

// NewDockerMirror creates a new Docker mirror handler
func NewDockerMirror(registries []string) *DockerMirror {
	return &DockerMirror{
		registries: registries,
	}
}

// getDockerConfigPath returns the path to Docker daemon config file
func (d *DockerMirror) getDockerConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// For Docker Desktop on macOS and Windows, use ~/.docker/daemon.json
	// For Linux, it's typically /etc/docker/daemon.json but we'll use user config
	// to avoid requiring sudo permissions
	if runtime.GOOS == "linux" {
		// Check if /etc/docker/daemon.json exists (system-wide config)
		systemPath := "/etc/docker/daemon.json"
		if _, err := os.Stat(systemPath); err == nil {
			// System config exists, but we can't modify it without sudo
			// Use user-level config instead
			return filepath.Join(homeDir, ".docker", "daemon.json"), nil
		}
	}

	return filepath.Join(homeDir, ".docker", "daemon.json"), nil
}

// isDockerDesktop checks if Docker Desktop is being used
func (d *DockerMirror) isDockerDesktop() bool {
	if runtime.GOOS == "darwin" {
		// Check if Docker Desktop is installed on macOS
		dockerDesktopPath := "/Applications/Docker.app"
		if _, err := os.Stat(dockerDesktopPath); err == nil {
			return true
		}
	}
	return false
}

// enableDockerDesktop provides instructions for Docker Desktop users
func (d *DockerMirror) enableDockerDesktop() error {
	fmt.Println("\n⚠ Docker Desktop detected!")
	fmt.Println("\nDocker Desktop doesn't use ~/.docker/daemon.json")
	fmt.Println("Please configure registry mirrors manually:")
	fmt.Println()
	fmt.Println("1. Open Docker Desktop")
	fmt.Println("2. Click Docker icon in menu bar → Settings")
	fmt.Println("3. Go to 'Docker Engine' tab")
	fmt.Println("4. Add the following to the JSON configuration:")
	fmt.Println()

	// Show registry mirrors if configured
	if len(d.registries) > 0 {
		fmt.Println("  \"registry-mirrors\": [")
		for i, reg := range d.registries {
			prefix := "https://"
			if !strings.HasPrefix(reg, "http://") && !strings.HasPrefix(reg, "https://") {
				reg = prefix + reg
			}
			if i < len(d.registries)-1 {
				fmt.Printf("    \"%s\",\n", reg)
			} else {
				fmt.Printf("    \"%s\"\n", reg)
			}
		}
		fmt.Println("  ]")
	}

	fmt.Println()
	fmt.Println("5. Click 'Apply & Restart'")
	fmt.Println()
	return nil
}

// Enable configures Docker to use registry mirrors
func (d *DockerMirror) Enable() error {
	// For Docker Desktop, provide instructions instead
	if d.isDockerDesktop() {
		return d.enableDockerDesktop()
	}

	configPath, err := d.getDockerConfigPath()
	if err != nil {
		return err
	}

	// Ensure .docker directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create .docker directory: %w", err)
	}

	// Read existing config or create new one
	var config map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new config
			config = make(map[string]interface{})
		} else {
			return fmt.Errorf("failed to read daemon.json: %w", err)
		}
	} else {
		// Parse existing config
		if err := json.Unmarshal(data, &config); err != nil {
			// Backup corrupted file
			backupPath := configPath + ".backup"
			os.WriteFile(backupPath, data, 0644)
			fmt.Printf("Warning: existing daemon.json is invalid, backed up to %s\n", backupPath)
			config = make(map[string]interface{})
		}
	}

	// Format registry URLs (ensure they have https:// prefix)
	if len(d.registries) > 0 {
		formattedRegistries := make([]string, len(d.registries))
		for i, reg := range d.registries {
			if !strings.HasPrefix(reg, "http://") && !strings.HasPrefix(reg, "https://") {
				formattedRegistries[i] = "https://" + reg
			} else {
				formattedRegistries[i] = reg
			}
		}
		// Set registry-mirrors
		config["registry-mirrors"] = formattedRegistries
	}

	// Write config back
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal daemon.json: %w", err)
	}

	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write daemon.json: %w", err)
	}

	return nil
}

// Disable removes registry mirror configuration
func (d *DockerMirror) Disable() error {
	// For Docker Desktop, provide instructions
	if d.isDockerDesktop() {
		fmt.Println("\n⚠ Docker Desktop detected!")
		fmt.Println("To disable registry mirrors:")
		fmt.Println("1. Open Docker Desktop → Settings → Docker Engine")
		fmt.Println("2. Remove the 'registry-mirrors' section")
		fmt.Println("3. Click 'Apply & Restart'")
		fmt.Println()
		return nil
	}

	configPath, err := d.getDockerConfigPath()
	if err != nil {
		return err
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to disable
		}
		return fmt.Errorf("failed to read daemon.json: %w", err)
	}

	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse daemon.json: %w", err)
	}

	// Remove registry-mirrors
	delete(config, "registry-mirrors")

	// If config is now empty, remove the file
	if len(config) == 0 {
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove daemon.json: %w", err)
		}
		return nil
	}

	// Write config back
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal daemon.json: %w", err)
	}

	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write daemon.json: %w", err)
	}

	return nil
}

// Status checks if registry mirrors are currently configured
func (d *DockerMirror) Status() (bool, string, error) {
	// For Docker Desktop, we can't easily read the config
	if d.isDockerDesktop() {
		return false, "check Docker Desktop settings", nil
	}

	configPath, err := d.getDockerConfigPath()
	if err != nil {
		return false, "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "default registry", nil
		}
		return false, "", fmt.Errorf("failed to read daemon.json: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return false, "", fmt.Errorf("failed to parse daemon.json: %w", err)
	}

	mirrors, ok := config["registry-mirrors"]
	if !ok {
		return false, "default registry", nil
	}

	// Convert mirrors to string representation
	mirrorsSlice, ok := mirrors.([]interface{})
	if !ok || len(mirrorsSlice) == 0 {
		return false, "default registry", nil
	}

	mirrorStrings := make([]string, 0, len(mirrorsSlice))
	for _, m := range mirrorsSlice {
		if str, ok := m.(string); ok {
			// Remove https:// prefix for cleaner display
			display := strings.TrimPrefix(str, "https://")
			display = strings.TrimPrefix(display, "http://")
			mirrorStrings = append(mirrorStrings, display)
		}
	}

	if len(mirrorStrings) == 0 {
		return false, "default registry", nil
	}

	return true, strings.Join(mirrorStrings, ", "), nil
}
