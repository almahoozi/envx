package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Manager handles configuration operations
type Manager struct {
	loader *Loader
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		loader: NewLoader(),
	}
}

// GetEffectiveConfig returns the effective configuration after merging all sources
func (m *Manager) GetEffectiveConfig() (*Config, error) {
	return m.loader.LoadEffectiveConfig()
}

// Set sets a configuration value in the global config
func (m *Manager) Set(key, value string, isDirectory bool) error {
	// Load existing config
	var existingConfig *Config
	var err error

	if isDirectory {
		existingConfig, err = m.loader.LoadDirectoryConfig()
	} else {
		existingConfig, err = m.loader.LoadGlobalConfig()
	}
	if err != nil {
		return err
	}

	// If no existing config, create a new one
	if existingConfig == nil {
		existingConfig = &Config{}
	}

	// Set the value
	if err := m.setConfigValue(existingConfig, key, value); err != nil {
		return err
	}

	// Save the config
	if isDirectory {
		return m.loader.SaveDirectoryConfig(existingConfig)
	}
	return m.loader.SaveGlobalConfig(existingConfig)
}

// Get gets a configuration value, returning the effective value and its source
func (m *Manager) Get(key string) (string, ConfigSource, error) {
	report, err := m.GetReport()
	if err != nil {
		return "", "", err
	}

	switch strings.ToLower(key) {
	case "keystore":
		return report.Keystore.Value, report.Keystore.Source, nil
	case "file":
		return report.File.Value, report.File.Source, nil
	case "format":
		return report.Format.Value, report.Format.Source, nil
	case "key_name", "keyname":
		return report.KeyName.Value, report.KeyName.Source, nil
	default:
		return "", "", fmt.Errorf("unknown configuration key: %s", key)
	}
}

// GetReport returns a detailed report of all configuration values with their sources
func (m *Manager) GetReport() (*ConfigReport, error) {
	defaults := DefaultConfig()

	globalConfig, err := m.loader.LoadGlobalConfig()
	if err != nil {
		return nil, err
	}

	dirConfig, err := m.loader.LoadDirectoryConfig()
	if err != nil {
		return nil, err
	}

	report := &ConfigReport{
		Keystore: ConfigValue{Value: defaults.Keystore, Source: SourceDefault},
		File:     ConfigValue{Value: defaults.File, Source: SourceDefault},
		Format:   ConfigValue{Value: defaults.Format, Source: SourceDefault},
		KeyName:  ConfigValue{Value: defaults.KeyName, Source: SourceDefault},
	}

	// Apply global config
	if globalConfig != nil {
		if globalConfig.Keystore != "" {
			report.Keystore = ConfigValue{Value: globalConfig.Keystore, Source: SourceGlobal}
		}
		if globalConfig.File != "" {
			report.File = ConfigValue{Value: globalConfig.File, Source: SourceGlobal}
		}
		if globalConfig.Format != "" {
			report.Format = ConfigValue{Value: globalConfig.Format, Source: SourceGlobal}
		}
		if globalConfig.KeyName != "" {
			report.KeyName = ConfigValue{Value: globalConfig.KeyName, Source: SourceGlobal}
		}
	}

	// Apply directory config (overrides global)
	if dirConfig != nil {
		if dirConfig.Keystore != "" {
			report.Keystore = ConfigValue{Value: dirConfig.Keystore, Source: SourceDirectory}
		}
		if dirConfig.File != "" {
			report.File = ConfigValue{Value: dirConfig.File, Source: SourceDirectory}
		}
		if dirConfig.Format != "" {
			report.Format = ConfigValue{Value: dirConfig.Format, Source: SourceDirectory}
		}
		if dirConfig.KeyName != "" {
			report.KeyName = ConfigValue{Value: dirConfig.KeyName, Source: SourceDirectory}
		}
	}

	return report, nil
}

// Reset removes a configuration key from the specified config file
func (m *Manager) Reset(key string, isDirectory bool) error {
	// Load existing config
	var existingConfig *Config
	var err error

	if isDirectory {
		existingConfig, err = m.loader.LoadDirectoryConfig()
	} else {
		existingConfig, err = m.loader.LoadGlobalConfig()
	}
	if err != nil {
		return err
	}

	// If no existing config, nothing to reset
	if existingConfig == nil {
		return nil
	}

	// Reset the value (set to empty string so it gets omitted in YAML)
	if err := m.setConfigValue(existingConfig, key, ""); err != nil {
		return err
	}

	// If all values are empty, delete the config file
	if m.isConfigEmpty(existingConfig) {
		if isDirectory {
			return m.loader.DeleteDirectoryConfig()
		}
		return m.loader.DeleteGlobalConfig()
	}

	// Save the updated config
	if isDirectory {
		return m.loader.SaveDirectoryConfig(existingConfig)
	}
	return m.loader.SaveGlobalConfig(existingConfig)
}

// Init initializes a configuration file (creates empty file or removes all entries)
func (m *Manager) Init(isDirectory bool) error {
	// Delete the config file to reset to defaults
	if isDirectory {
		return m.loader.DeleteDirectoryConfig()
	}
	return m.loader.DeleteGlobalConfig()
}

// GetConfigPaths returns the paths to the global and directory config files
func (m *Manager) GetConfigPaths() (string, string, error) {
	globalPath, err := m.loader.GetGlobalConfigPath()
	if err != nil {
		return "", "", err
	}

	dirPath, err := m.loader.GetDirectoryConfigPath()
	if err != nil {
		return "", "", err
	}

	return globalPath, dirPath, nil
}

// ConfigExists returns whether the global and directory config files exist
func (m *Manager) ConfigExists() (bool, bool, error) {
	globalPath, dirPath, err := m.GetConfigPaths()
	if err != nil {
		return false, false, err
	}

	globalExists := false
	if _, err := os.Stat(globalPath); err == nil {
		globalExists = true
	}

	dirExists := false
	if _, err := os.Stat(dirPath); err == nil {
		dirExists = true
	}

	return globalExists, dirExists, nil
}

// setConfigValue sets a configuration value using reflection
func (m *Manager) setConfigValue(config *Config, key, value string) error {
	v := reflect.ValueOf(config).Elem()

	switch strings.ToLower(key) {
	case "keystore":
		v.FieldByName("Keystore").SetString(value)
	case "file":
		v.FieldByName("File").SetString(value)
	case "format":
		v.FieldByName("Format").SetString(value)
	case "key_name", "keyname":
		v.FieldByName("KeyName").SetString(value)
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}

// isConfigEmpty checks if all configuration values are empty
func (m *Manager) isConfigEmpty(config *Config) bool {
	return config.Keystore == "" && config.File == "" && config.Format == "" && config.KeyName == ""
}
