package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the crosh configuration structure
type Config struct {
	Mirror MirrorConfig `yaml:"mirror"`
	Proxy  ProxyConfig  `yaml:"proxy"`
}

// MirrorConfig contains mirror settings for package managers
type MirrorConfig struct {
	NPM     string   `yaml:"npm"`
	Pip     string   `yaml:"pip"`
	Apt     string   `yaml:"apt"`
	Cargo   string   `yaml:"cargo"`
	Go      string   `yaml:"go"`
	Docker  []string `yaml:"docker"`
	Enabled bool     `yaml:"enabled"`
}

// ProxyConfig contains proxy settings
type ProxyConfig struct {
	SubscriptionURL string `yaml:"subscription_url"`
	LocalPort       int    `yaml:"local_port"`
	Enabled         bool   `yaml:"enabled"`
	XrayPath        string `yaml:"xray_path"`
	CurrentNode     string `yaml:"current_node,omitempty"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Mirror: MirrorConfig{
			NPM:   "https://registry.npmmirror.com",
			Pip:   "https://mirrors.aliyun.com/pypi/simple/",
			Apt:   "mirrors.aliyun.com",
			Cargo: "https://mirrors.ustc.edu.cn/crates.io-index",
			Go:    "https://goproxy.cn,direct",
			Docker: []string{
				"docker.1ms.run",
				"docker.m.daocloud.io",
			},
			Enabled: false,
		},
		Proxy: ProxyConfig{
			SubscriptionURL: "",
			LocalPort:       7676,
			Enabled:         false,
			XrayPath:        filepath.Join(homeDir, ".crosh", "xray-core"),
		},
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".crosh")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// Load reads the configuration from the config file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save writes the configuration to the config file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
