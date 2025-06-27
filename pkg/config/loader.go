package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader handles loading configuration from files
type Loader struct{}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadEffectiveConfig loads and merges configuration from all sources
func (l *Loader) LoadEffectiveConfig() (*Config, error) {
	config := DefaultConfig()

	// Load global config
	globalConfig, err := l.LoadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}
	if globalConfig != nil {
		config.Merge(globalConfig)
	}

	// Load directory config
	dirConfig, err := l.LoadDirectoryConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load directory config: %w", err)
	}
	if dirConfig != nil {
		config.Merge(dirConfig)
	}

	return config, nil
}

// LoadGlobalConfig loads the global configuration file
func (l *Loader) LoadGlobalConfig() (*Config, error) {
	configPath, err := l.GetGlobalConfigPath()
	if err != nil {
		return nil, err
	}

	return l.loadConfigFromFile(configPath)
}

// LoadDirectoryConfig loads the directory-specific configuration file
func (l *Loader) LoadDirectoryConfig() (*Config, error) {
	configPath, err := l.GetDirectoryConfigPath()
	if err != nil {
		return nil, err
	}

	return l.loadConfigFromFile(configPath)
}

// GetGlobalConfigPath returns the path to the global configuration file
func (l *Loader) GetGlobalConfigPath() (string, error) {
	var configDir string
	var err error

	// Check for override environment variable first (for testing safety)
	if envConfigDir := os.Getenv("ENVX_CONFIG_DIR"); envConfigDir != "" {
		configDir = envConfigDir
	} else {
		configDir, err = os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user config directory: %w", err)
		}
		configDir = filepath.Join(configDir, "envx")
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// GetDirectoryConfigPath returns the path to the directory-specific configuration file
func (l *Loader) GetDirectoryConfigPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return filepath.Join(wd, ".envx.yaml"), nil
}

// loadConfigFromFile loads configuration from a specific file
func (l *Loader) loadConfigFromFile(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil // File doesn't exist, return nil config
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return &config, nil
}

// SaveGlobalConfig saves the global configuration file
func (l *Loader) SaveGlobalConfig(config *Config) error {
	configPath, err := l.GetGlobalConfigPath()
	if err != nil {
		return err
	}

	return l.saveConfigToFile(configPath, config)
}

// SaveDirectoryConfig saves the directory-specific configuration file
func (l *Loader) SaveDirectoryConfig(config *Config) error {
	configPath, err := l.GetDirectoryConfigPath()
	if err != nil {
		return err
	}

	return l.saveConfigToFile(configPath, config)
}

// saveConfigToFile saves configuration to a specific file
func (l *Loader) saveConfigToFile(path string, config *Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}

	return nil
}

// DeleteGlobalConfig deletes the global configuration file
func (l *Loader) DeleteGlobalConfig() error {
	configPath, err := l.GetGlobalConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(configPath)
}

// DeleteDirectoryConfig deletes the directory-specific configuration file
func (l *Loader) DeleteDirectoryConfig() error {
	configPath, err := l.GetDirectoryConfigPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(configPath)
}
