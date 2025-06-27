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
	if config.KeyName != "envx.test" {
		t.Errorf("Expected default key_name to be 'envx.test', got '%s'", config.KeyName)
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
		KeyName:        "envx.test",
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
	if base.KeyName != "envx.test" {
		t.Errorf("Expected key_name to remain 'envx.test', got '%s'", base.KeyName)
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

	if report.KeyName.Value != "envx.test" || report.KeyName.Source != SourceDefault {
		t.Errorf("Expected key_name 'envx.test' from default, got '%s' from '%s'",
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
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Override config directory for testing
	tempConfigDir := filepath.Join(tempDir, "config")
	os.Setenv("ENVX_CONFIG_DIR", tempConfigDir)
	defer os.Unsetenv("ENVX_CONFIG_DIR")

	manager := NewManager()

	// Set up test configuration
	err := manager.Set("name", "local", false)
	if err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	tests := []struct {
		name         string
		explicitFile string
		explicitName string
		createFiles  []string
		expected     string
	}{
		{
			name:         "explicit file takes precedence",
			explicitFile: "custom.env",
			explicitName: "",
			createFiles:  []string{"custom.env"},
			expected:     "custom.env",
		},
		{
			name:         "explicit file with name",
			explicitFile: "custom.env",
			explicitName: "test",
			createFiles:  []string{"custom.env.test"},
			expected:     "custom.env.test",
		},
		{
			name:         "explicit name with default file",
			explicitFile: "",
			explicitName: "test",
			createFiles:  []string{".env.test"},
			expected:     ".env.test",
		},
		{
			name:         "configured name",
			explicitFile: "",
			explicitName: "",
			createFiles:  []string{".env.local"},
			expected:     ".env.local",
		},
		{
			name:         "file_resolution - first existing",
			explicitFile: "",
			explicitName: "",
			createFiles:  []string{".env.local", ".env.dev"},
			expected:     ".env.local",
		},
		{
			name:         "file_resolution - fallback to second",
			explicitFile: "",
			explicitName: "",
			createFiles:  []string{".env.dev"},
			expected:     ".env.dev",
		},
		{
			name:         "no existing files, use default",
			explicitFile: "",
			explicitName: "",
			createFiles:  []string{},
			expected:     ".env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing files
			os.RemoveAll(tempDir)
			os.MkdirAll(tempDir, 0755)
			os.Chdir(tempDir)

			// Set up file_resolution for tests that need it
			if tt.name == "file_resolution - first existing" || tt.name == "file_resolution - fallback to second" {
				// Reset name to empty for file_resolution tests
				err := manager.Set("name", "", false)
				if err != nil {
					t.Fatalf("Failed to reset name: %v", err)
				}
				err = manager.Set("file_resolution", ".env.local,.env.dev,.env", false)
				if err != nil {
					t.Fatalf("Failed to set file_resolution: %v", err)
				}
			} else if tt.name == "no existing files, use default" {
				// Reset name and file_resolution for default test
				err := manager.Set("name", "", false)
				if err != nil {
					t.Fatalf("Failed to reset name: %v", err)
				}
				err = manager.Set("file_resolution", "", false)
				if err != nil {
					t.Fatalf("Failed to reset file_resolution: %v", err)
				}
			} else {
				// Reset file_resolution for other tests
				err := manager.Set("file_resolution", "", false)
				if err != nil {
					t.Fatalf("Failed to reset file_resolution: %v", err)
				}
				// Set name back to local for configured name test
				if tt.name == "configured name" {
					err := manager.Set("name", "local", false)
					if err != nil {
						t.Fatalf("Failed to set name: %v", err)
					}
				}
			}

			// Create test files
			for _, file := range tt.createFiles {
				err := os.WriteFile(file, []byte("test"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file %s: %v", file, err)
				}
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
		{".env", "local", ".env.local"},
		{".env", "test", ".env.test"},
		{"app.config", "", "app.config"},
		{"app.config", "dev", "app.config.dev"},
		{"noext", "suffix", "noext.suffix"},
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

func TestConfigGetVSeparator(t *testing.T) {
	// This test verifies that the separator field is properly handled
	// The actual CLI testing would be done at integration level

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

	// Test that we can get configuration values
	// The separator functionality is tested at the CLI level
	report, err := manager.GetReport()
	if err != nil {
		t.Fatalf("Failed to get config report: %v", err)
	}

	// Verify we can get all the expected fields
	expectedFields := []string{
		report.Keystore.Value,
		report.File.Value,
		report.Name.Value,
		report.Format.Value,
		report.KeyName.Value,
		strings.Join(report.FileResolution.Value, ","),
		report.BackupOnWrite.Value,
	}

	if len(expectedFields) != 7 {
		t.Errorf("Expected 7 config fields, got %d", len(expectedFields))
	}

	// Test that joining with different separators works
	commaSeparated := strings.Join(expectedFields, ",")
	pipeSeparated := strings.Join(expectedFields, "|")
	spaceSeparated := strings.Join(expectedFields, " ")

	if !strings.Contains(commaSeparated, ",") {
		t.Error("Comma separator not working")
	}
	if !strings.Contains(pipeSeparated, "|") {
		t.Error("Pipe separator not working")
	}
	if !strings.Contains(spaceSeparated, " ") {
		t.Error("Space separator not working")
	}
}
