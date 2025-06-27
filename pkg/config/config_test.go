package config

import (
	"os"
	"path/filepath"
	"strings"
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
	if config.Name != "" {
		t.Errorf("Expected default name to be empty, got '%s'", config.Name)
	}
	if config.Format != "env" {
		t.Errorf("Expected default format to be 'env', got '%s'", config.Format)
	}
	if config.KeyName != "default" {
		t.Errorf("Expected default key_name to be 'default', got '%s'", config.KeyName)
	}
	if len(config.FileResolution) != 1 || config.FileResolution[0] != ".env" {
		t.Errorf("Expected default file_resolution to be ['.env'], got %v", config.FileResolution)
	}
	if !config.BackupOnWrite {
		t.Errorf("Expected default backup_on_write to be true, got %v", config.BackupOnWrite)
	}
}

func TestConfigMerge(t *testing.T) {
	base := &Config{
		Keystore:       "macos",
		File:           ".env",
		Name:           "",
		Format:         "env",
		KeyName:        "default",
		FileResolution: []string{".env"},
		BackupOnWrite:  true,
	}

	override := &Config{
		Keystore:       "password",
		Format:         "json",
		Name:           "local",
		FileResolution: []string{".env.local", ".env"},
		BackupOnWrite:  false,
		// File and KeyName are empty, should not override
	}

	base.Merge(override)

	if base.Keystore != "password" {
		t.Errorf("Expected keystore to be overridden to 'password', got '%s'", base.Keystore)
	}
	if base.Format != "json" {
		t.Errorf("Expected format to be overridden to 'json', got '%s'", base.Format)
	}
	if base.Name != "local" {
		t.Errorf("Expected name to be overridden to 'local', got '%s'", base.Name)
	}
	if base.File != ".env" {
		t.Errorf("Expected file to remain '.env', got '%s'", base.File)
	}
	if base.KeyName != "default" {
		t.Errorf("Expected key_name to remain 'default', got '%s'", base.KeyName)
	}
	if len(base.FileResolution) != 2 || base.FileResolution[0] != ".env.local" {
		t.Errorf("Expected file_resolution to be overridden to ['.env.local', '.env'], got %v", base.FileResolution)
	}
	if base.BackupOnWrite != false {
		t.Errorf("Expected backup_on_write to be overridden to false, got %v", base.BackupOnWrite)
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

func TestManagerNewFields(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Override config directory for testing
	tempConfigDir := filepath.Join(tempDir, "config")
	os.Setenv("ENVX_CONFIG_DIR", tempConfigDir)
	defer os.Unsetenv("ENVX_CONFIG_DIR")

	manager := NewManager()

	// Test setting name
	err := manager.Set("name", "local", false)
	if err != nil {
		t.Fatalf("Failed to set name config: %v", err)
	}

	value, source, err := manager.Get("name")
	if err != nil {
		t.Fatalf("Failed to get name config: %v", err)
	}
	if value != "local" || source != SourceGlobal {
		t.Errorf("Expected name 'local' from global, got '%s' from '%s'", value, source)
	}

	// Test setting file_resolution
	err = manager.Set("file_resolution", ".env.local,.env,.env.example", false)
	if err != nil {
		t.Fatalf("Failed to set file_resolution config: %v", err)
	}

	values, source, err := manager.GetArray("file_resolution")
	if err != nil {
		t.Fatalf("Failed to get file_resolution config: %v", err)
	}
	expected := []string{".env.local", ".env", ".env.example"}
	if len(values) != len(expected) {
		t.Errorf("Expected file_resolution length %d, got %d", len(expected), len(values))
	}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("Expected file_resolution[%d] to be '%s', got '%s'", i, expected[i], v)
		}
	}
	if source != SourceGlobal {
		t.Errorf("Expected source to be global, got '%s'", source)
	}
}

func TestResolveFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Override config directory for testing
	tempConfigDir := filepath.Join(tempDir, "config")
	os.Setenv("ENVX_CONFIG_DIR", tempConfigDir)
	defer os.Unsetenv("ENVX_CONFIG_DIR")

	manager := NewManager()

	// Create some test files
	os.WriteFile(".env.local", []byte("TEST=local"), 0644)
	os.WriteFile(".env", []byte("TEST=default"), 0644)

	tests := []struct {
		name         string
		explicitFile string
		explicitName string
		configName   string
		fileRes      []string
		expected     string
	}{
		{
			name:         "explicit file takes precedence",
			explicitFile: "custom.env",
			explicitName: "",
			expected:     "custom.env",
		},
		{
			name:         "explicit file with name",
			explicitFile: "custom.env",
			explicitName: "test",
			expected:     "custom.test.env",
		},
		{
			name:         "explicit name with default file",
			explicitFile: "",
			explicitName: "test",
			expected:     ".test.env",
		},
		{
			name:       "configured name",
			configName: "local",
			expected:   ".local.env",
		},
		{
			name:     "file resolution - first existing",
			fileRes:  []string{".env.nonexistent", ".env.local", ".env"},
			expected: ".env.local",
		},
		{
			name:     "file resolution - fallback to second",
			fileRes:  []string{".env.nonexistent", ".env"},
			expected: ".env",
		},
		{
			name:     "no existing files, use default",
			fileRes:  []string{".env.nonexistent"},
			expected: ".env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset config
			manager.Init(false)

			// Set up config for this test
			if tt.configName != "" {
				manager.Set("name", tt.configName, false)
			}
			if len(tt.fileRes) > 0 {
				manager.Set("file_resolution", strings.Join(tt.fileRes, ","), false)
			}

			result, err := manager.ResolveFile(tt.explicitFile, tt.explicitName)
			if err != nil {
				t.Fatalf("ResolveFile failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildFilename(t *testing.T) {
	tests := []struct {
		file     string
		name     string
		expected string
	}{
		{".env", "", ".env"},
		{".env", "local", ".local.env"},
		{".env", "test", ".test.env"},
		{"config", "prod", "config.prod"},
		{"app.config", "dev", "app.dev.config"},
	}

	for _, tt := range tests {
		result := buildFilename(tt.file, tt.name)
		if result != tt.expected {
			t.Errorf("buildFilename(%s, %s) = %s, expected %s", tt.file, tt.name, result, tt.expected)
		}
	}
}

func TestEnvConfigDirOverride(t *testing.T) {
	// Set custom config directory
	customDir := "/tmp/envx-test-config"
	os.Setenv("ENVX_CONFIG_DIR", customDir)
	defer os.Unsetenv("ENVX_CONFIG_DIR")

	loader := NewLoader()
	path, err := loader.GetGlobalConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	expected := filepath.Join(customDir, "config.yaml")
	if path != expected {
		t.Errorf("Expected config path '%s', got '%s'", expected, path)
	}
}

func TestBackupOnWrite(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Override config directory for testing
	tempConfigDir := filepath.Join(tempDir, "config")
	os.Setenv("ENVX_CONFIG_DIR", tempConfigDir)
	defer os.Unsetenv("ENVX_CONFIG_DIR")

	manager := NewManager()

	// Test setting backup_on_write
	err := manager.Set("backup_on_write", "false", false)
	if err != nil {
		t.Fatalf("Failed to set backup_on_write config: %v", err)
	}

	value, source, err := manager.Get("backup_on_write")
	if err != nil {
		t.Fatalf("Failed to get backup_on_write config: %v", err)
	}
	if value != "false" || source != SourceGlobal {
		t.Errorf("Expected backup_on_write 'false' from global, got '%s' from '%s'", value, source)
	}

	// Test boolean parsing
	err = manager.Set("backup_on_write", "true", false)
	if err != nil {
		t.Fatalf("Failed to set backup_on_write to true: %v", err)
	}

	value, _, err = manager.Get("backup_on_write")
	if err != nil {
		t.Fatalf("Failed to get backup_on_write config: %v", err)
	}
	if value != "true" {
		t.Errorf("Expected backup_on_write 'true', got '%s'", value)
	}
}

func TestCreateBackup(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Override config directory for testing
	tempConfigDir := filepath.Join(tempDir, "config")
	os.Setenv("ENVX_CONFIG_DIR", tempConfigDir)
	defer os.Unsetenv("ENVX_CONFIG_DIR")

	manager := NewManager()

	// Create a test file
	testFile := "test-config.yaml"
	originalContent := "test: original"
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Enable backup
	err = manager.Set("backup_on_write", "true", false)
	if err != nil {
		t.Fatalf("Failed to enable backup: %v", err)
	}

	// Create backup
	err = manager.createBackup(testFile)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Check that backup file was created
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	backupFound := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "test-config.yaml.backup.") {
			backupFound = true

			// Verify backup content
			backupContent, err := os.ReadFile(file.Name())
			if err != nil {
				t.Fatalf("Failed to read backup file: %v", err)
			}

			if string(backupContent) != originalContent {
				t.Errorf("Backup content doesn't match original. Expected '%s', got '%s'", originalContent, string(backupContent))
			}
			break
		}
	}

	if !backupFound {
		t.Error("Backup file was not created")
	}

	// Test with backup disabled
	err = manager.Set("backup_on_write", "false", false)
	if err != nil {
		t.Fatalf("Failed to disable backup: %v", err)
	}

	// Try to create backup (should be skipped)
	testFile2 := "test-config2.yaml"
	err = os.WriteFile(testFile2, []byte("test: content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	err = manager.createBackup(testFile2)
	if err != nil {
		t.Fatalf("Failed to create backup (should be skipped): %v", err)
	}

	// Check that no backup was created for testFile2
	files, err = os.ReadDir(".")
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "test-config2.yaml.backup.") {
			t.Error("Backup file was created when backup was disabled")
		}
	}
}
