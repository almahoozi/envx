package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Keystore != "macos" {
		t.Errorf("Expected default keystore to be 'macos', got '%s'", config.Keystore)
	}
	if config.File != ".env" {
		t.Errorf("Expected default file to be '.env', got '%s'", config.File)
	}
	if config.Format != "env" {
		t.Errorf("Expected default format to be 'env', got '%s'", config.Format)
	}
	if config.KeyName != "default" {
		t.Errorf("Expected default key_name to be 'default', got '%s'", config.KeyName)
	}
}

func TestConfigMerge(t *testing.T) {
	base := &Config{
		Keystore: "macos",
		File:     ".env",
		Format:   "env",
		KeyName:  "default",
	}

	override := &Config{
		Keystore: "password",
		Format:   "json",
		// File and KeyName are empty, should not override
	}

	base.Merge(override)

	if base.Keystore != "password" {
		t.Errorf("Expected keystore to be overridden to 'password', got '%s'", base.Keystore)
	}
	if base.Format != "json" {
		t.Errorf("Expected format to be overridden to 'json', got '%s'", base.Format)
	}
	if base.File != ".env" {
		t.Errorf("Expected file to remain '.env', got '%s'", base.File)
	}
	if base.KeyName != "default" {
		t.Errorf("Expected key_name to remain 'default', got '%s'", base.KeyName)
	}
}

func TestManagerSetAndGet(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Override config directory for testing
	tempConfigDir := filepath.Join(tempDir, "config")
	os.Setenv("XDG_CONFIG_HOME", tempConfigDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	manager := NewManager()

	// Test setting and getting global config
	err := manager.Set("keystore", "password", false)
	if err != nil {
		t.Fatalf("Failed to set global config: %v", err)
	}

	value, source, err := manager.Get("keystore")
	if err != nil {
		t.Fatalf("Failed to get config value: %v", err)
	}

	if value != "password" {
		t.Errorf("Expected keystore value to be 'password', got '%s'", value)
	}
	if source != SourceGlobal {
		t.Errorf("Expected source to be global, got '%s'", source)
	}

	// Test setting directory config (should override global)
	err = manager.Set("keystore", "mock", true)
	if err != nil {
		t.Fatalf("Failed to set directory config: %v", err)
	}

	value, source, err = manager.Get("keystore")
	if err != nil {
		t.Fatalf("Failed to get config value: %v", err)
	}

	if value != "mock" {
		t.Errorf("Expected keystore value to be 'mock', got '%s'", value)
	}
	if source != SourceDirectory {
		t.Errorf("Expected source to be directory, got '%s'", source)
	}
}

func TestManagerReset(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Override config directory for testing
	tempConfigDir := filepath.Join(tempDir, "config")
	os.Setenv("XDG_CONFIG_HOME", tempConfigDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	manager := NewManager()

	// Set a value
	err := manager.Set("keystore", "password", false)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// Verify it's set
	value, source, err := manager.Get("keystore")
	if err != nil {
		t.Fatalf("Failed to get config value: %v", err)
	}
	if value != "password" || source != SourceGlobal {
		t.Fatalf("Config not set correctly")
	}

	// Reset the value
	err = manager.Reset("keystore", false)
	if err != nil {
		t.Fatalf("Failed to reset config: %v", err)
	}

	// Verify it's back to default
	value, source, err = manager.Get("keystore")
	if err != nil {
		t.Fatalf("Failed to get config value: %v", err)
	}
	if value != "macos" || source != SourceDefault {
		t.Errorf("Expected keystore to be reset to default 'macos', got '%s' from '%s'", value, source)
	}
}

func TestManagerInit(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Override config directory for testing
	tempConfigDir := filepath.Join(tempDir, "config")
	os.Setenv("XDG_CONFIG_HOME", tempConfigDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	manager := NewManager()

	// Set some values
	err := manager.Set("keystore", "password", false)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	// Verify config exists
	globalExists, _, err := manager.ConfigExists()
	if err != nil {
		t.Fatalf("Failed to check config existence: %v", err)
	}
	if !globalExists {
		t.Fatal("Expected global config to exist")
	}

	// Initialize (should delete the config)
	err = manager.Init(false)
	if err != nil {
		t.Fatalf("Failed to init config: %v", err)
	}

	// Verify config no longer exists
	globalExists, _, err = manager.ConfigExists()
	if err != nil {
		t.Fatalf("Failed to check config existence: %v", err)
	}
	if globalExists {
		t.Error("Expected global config to not exist after init")
	}

	// Verify values are back to defaults
	value, source, err := manager.Get("keystore")
	if err != nil {
		t.Fatalf("Failed to get config value: %v", err)
	}
	if value != "macos" || source != SourceDefault {
		t.Errorf("Expected keystore to be default 'macos', got '%s' from '%s'", value, source)
	}
}

func TestManagerGetReport(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Override config directory for testing
	tempConfigDir := filepath.Join(tempDir, "config")
	os.Setenv("XDG_CONFIG_HOME", tempConfigDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	manager := NewManager()

	// Set global and directory configs
	err := manager.Set("keystore", "password", false) // global
	if err != nil {
		t.Fatalf("Failed to set global config: %v", err)
	}

	err = manager.Set("format", "json", true) // directory
	if err != nil {
		t.Fatalf("Failed to set directory config: %v", err)
	}

	// Get report
	report, err := manager.GetReport()
	if err != nil {
		t.Fatalf("Failed to get config report: %v", err)
	}

	// Verify sources
	if report.Keystore.Value != "password" || report.Keystore.Source != SourceGlobal {
		t.Errorf("Expected keystore 'password' from global, got '%s' from '%s'",
			report.Keystore.Value, report.Keystore.Source)
	}

	if report.Format.Value != "json" || report.Format.Source != SourceDirectory {
		t.Errorf("Expected format 'json' from directory, got '%s' from '%s'",
			report.Format.Value, report.Format.Source)
	}

	if report.File.Value != ".env" || report.File.Source != SourceDefault {
		t.Errorf("Expected file '.env' from default, got '%s' from '%s'",
			report.File.Value, report.File.Source)
	}

	if report.KeyName.Value != "default" || report.KeyName.Source != SourceDefault {
		t.Errorf("Expected key_name 'default' from default, got '%s' from '%s'",
			report.KeyName.Value, report.KeyName.Source)
	}
}

func TestInvalidConfigKey(t *testing.T) {
	manager := NewManager()

	err := manager.Set("invalid_key", "value", false)
	if err == nil {
		t.Error("Expected error when setting invalid config key")
	}

	_, _, err = manager.Get("invalid_key")
	if err == nil {
		t.Error("Expected error when getting invalid config key")
	}
}
