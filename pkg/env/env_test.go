package env

import (
	"context"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/almahoozi/envx/pkg/crypto"
)

func TestVariable(t *testing.T) {
	v := Variable{Key: "TEST_KEY", Value: "test_value"}

	if v.Key != "TEST_KEY" {
		t.Errorf("Variable.Key = %q, want %q", v.Key, "TEST_KEY")
	}

	if v.Value != "test_value" {
		t.Errorf("Variable.Value = %q, want %q", v.Value, "test_value")
	}
}

func TestVariables_ToMap(t *testing.T) {
	vars := Variables{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "value2"},
		{Key: "KEY3", Value: "value3"},
	}

	m := vars.ToMap()

	expected := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "value3",
	}

	if len(m) != len(expected) {
		t.Errorf("ToMap() returned map of length %d, want %d", len(m), len(expected))
	}

	for k, v := range expected {
		if m[k] != v {
			t.Errorf("ToMap()[%q] = %q, want %q", k, m[k], v)
		}
	}
}

func TestVariables_Get(t *testing.T) {
	vars := Variables{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "value2"},
	}

	// Test existing key
	v := vars.Get("KEY1")
	if v == nil {
		t.Fatal("Get() returned nil for existing key")
	}
	if v.Key != "KEY1" || v.Value != "value1" {
		t.Errorf("Get() = {%q, %q}, want {%q, %q}", v.Key, v.Value, "KEY1", "value1")
	}

	// Test non-existing key
	v = vars.Get("NONEXISTENT")
	if v != nil {
		t.Errorf("Get() returned non-nil for non-existing key: %+v", v)
	}
}

func TestVariables_Set(t *testing.T) {
	vars := Variables{
		{Key: "EXISTING", Value: "old_value"},
	}

	// Test updating existing key
	vars.Set("EXISTING", "new_value")
	v := vars.Get("EXISTING")
	if v == nil || v.Value != "new_value" {
		t.Errorf("Set() didn't update existing key correctly")
	}

	// Test adding new key
	vars.Set("NEW_KEY", "new_value")
	v = vars.Get("NEW_KEY")
	if v == nil || v.Value != "new_value" {
		t.Errorf("Set() didn't add new key correctly")
	}

	if len(vars) != 2 {
		t.Errorf("Set() resulted in wrong number of variables: got %d, want 2", len(vars))
	}
}

func TestVariables_Remove(t *testing.T) {
	vars := Variables{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "value2"},
		{Key: "KEY3", Value: "value3"},
	}

	// Test removing existing key
	removed := vars.Remove("KEY2")
	if !removed {
		t.Error("Remove() returned false for existing key")
	}

	if len(vars) != 2 {
		t.Errorf("Remove() left wrong number of variables: got %d, want 2", len(vars))
	}

	if vars.Get("KEY2") != nil {
		t.Error("Remove() didn't actually remove the key")
	}

	// Verify remaining keys
	if vars.Get("KEY1") == nil || vars.Get("KEY3") == nil {
		t.Error("Remove() affected other keys")
	}

	// Test removing non-existing key
	removed = vars.Remove("NONEXISTENT")
	if removed {
		t.Error("Remove() returned true for non-existing key")
	}

	if len(vars) != 2 {
		t.Error("Remove() changed variable count for non-existing key")
	}
}

func TestBuildFilename(t *testing.T) {
	tests := []struct {
		baseFile string
		name     string
		expected string
	}{
		{".env", "", ".env"},
		{".env", "local", ".env.local"},
		{".env", "production", ".env.production"},
		{"custom.env", "", "custom.env"},
		{"custom.env", "test", "custom.env.test"},
	}

	for _, tt := range tests {
		t.Run(tt.baseFile+"_"+tt.name, func(t *testing.T) {
			result := BuildFilename(tt.baseFile, tt.name)
			if result != tt.expected {
				t.Errorf("BuildFilename(%q, %q) = %q, want %q", tt.baseFile, tt.name, result, tt.expected)
			}
		})
	}
}

func TestFileLoader_Load(t *testing.T) {
	loader := NewFileLoader()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		content  string
		expected Variables
		wantErr  bool
	}{
		{
			name: "valid env file",
			content: `KEY1=value1
KEY2=value2
KEY3="quoted value"`,
			expected: Variables{
				{Key: "KEY1", Value: "value1"},
				{Key: "KEY2", Value: "value2"},
				{Key: "KEY3", Value: "quoted value"},
			},
			wantErr: false,
		},
		{
			name: "file with comments and empty lines",
			content: `# This is a comment
KEY1=value1

# Another comment
KEY2=value2
`,
			expected: Variables{
				{Key: "KEY1", Value: "value1"},
				{Key: "KEY2", Value: "value2"},
			},
			wantErr: false,
		},
		{
			name: "file with whitespace",
			content: `  KEY1  =  value1  
	KEY2	=	value2	
KEY3= value3 `,
			expected: Variables{
				{Key: "KEY1", Value: "value1"},
				{Key: "KEY2", Value: "value2"},
				{Key: "KEY3", Value: "value3"},
			},
			wantErr: false,
		},
		{
			name:     "empty file",
			content:  "",
			expected: Variables{},
			wantErr:  false,
		},
		{
			name: "malformed lines",
			content: `KEY1=value1
MALFORMED_LINE_NO_EQUALS
KEY2=value2
=value_without_key`,
			expected: Variables{
				{Key: "KEY1", Value: "value1"},
				{Key: "KEY2", Value: "value2"},
			},
			wantErr: false,
		},
		{
			name: "special characters in values",
			content: `KEY1=value with spaces
KEY2="value with = equals"
KEY3=value!@#$%^&*()`,
			expected: Variables{
				{Key: "KEY1", Value: "value with spaces"},
				{Key: "KEY2", Value: "value with = equals"},
				{Key: "KEY3", Value: "value!@#$%^&*()"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filename := filepath.Join(tempDir, tt.name+".env")
			err := os.WriteFile(filename, []byte(tt.content), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// Test loading
			vars, err := loader.Load(context.Background(), filename)

			if tt.wantErr {
				if err == nil {
					t.Error("Load() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error: %v", err)
				return
			}

			if len(vars) != len(tt.expected) {
				t.Errorf("Load() returned %d variables, want %d", len(vars), len(tt.expected))
			}

			for i, expected := range tt.expected {
				if i >= len(vars) {
					t.Errorf("Load() missing variable %d: %+v", i, expected)
					continue
				}

				if vars[i].Key != expected.Key || vars[i].Value != expected.Value {
					t.Errorf("Load() variable %d = {%q, %q}, want {%q, %q}",
						i, vars[i].Key, vars[i].Value, expected.Key, expected.Value)
				}
			}
		})
	}

	// Test non-existent file
	vars, err := loader.Load(context.Background(), filepath.Join(tempDir, "nonexistent.env"))
	if err != nil {
		t.Errorf("Load() non-existent file should not error: %v", err)
	}
	if len(vars) != 0 {
		t.Errorf("Load() non-existent file should return empty variables, got %d", len(vars))
	}
}

func TestFileLoader_LoadWithDecryption(t *testing.T) {
	loader := NewFileLoader()
	encryptor := crypto.NewAESEncryptor()

	// Generate a test key
	key := make([]byte, crypto.KeySize)
	rand.Read(key)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test content with encrypted values
	plainValues := map[string]string{
		"SECRET1": "secret_value_1",
		"SECRET2": "secret_value_2",
		"PLAIN":   "plain_value",
	}

	var contentLines []string
	for k, v := range plainValues {
		if strings.HasPrefix(k, "SECRET") {
			encrypted, err := encryptor.Encrypt(v, key)
			if err != nil {
				t.Fatal(err)
			}
			contentLines = append(contentLines, k+"="+encrypted)
		} else {
			contentLines = append(contentLines, k+"="+v)
		}
	}

	content := strings.Join(contentLines, "\n")
	filename := filepath.Join(tempDir, "encrypted.env")
	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test loading with decryption
	vars, err := loader.LoadWithDecryption(context.Background(), filename, encryptor, key)
	if err != nil {
		t.Errorf("LoadWithDecryption() unexpected error: %v", err)
	}

	// Verify all values are decrypted
	varMap := vars.ToMap()
	for k, expectedValue := range plainValues {
		if varMap[k] != expectedValue {
			t.Errorf("LoadWithDecryption() %s = %q, want %q", k, varMap[k], expectedValue)
		}
	}
}

func TestFileWriter_Write(t *testing.T) {
	writer := NewFileWriter()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	vars := Variables{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "value with spaces"},
		{Key: "KEY3", Value: "value\"with\"quotes"},
	}

	tests := []struct {
		name         string
		format       Format
		expectedEnv  string
		expectedJSON string
		wantErr      bool
	}{
		{
			name:   "env format",
			format: FormatEnv,
			expectedEnv: `KEY1=value1
KEY2="value with spaces"
KEY3="value\"with\"quotes"
`,
			wantErr: false,
		},
		{
			name:         "json format",
			format:       FormatJSON,
			expectedJSON: `{"KEY1":"value1","KEY2":"value with spaces","KEY3":"value\"with\"quotes"}`,
			wantErr:      false,
		},
		{
			name:    "yaml format - not implemented",
			format:  FormatYAML,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(tempDir, tt.name+".test")

			err := writer.Write(filename, vars, tt.format)

			if tt.wantErr {
				if err == nil {
					t.Error("Write() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Write() unexpected error: %v", err)
				return
			}

			// Read the file back and verify content
			content, err := os.ReadFile(filename)
			if err != nil {
				t.Errorf("Failed to read written file: %v", err)
				return
			}

			contentStr := string(content)

			switch tt.format {
			case FormatEnv:
				if contentStr != tt.expectedEnv {
					t.Errorf("Write() env format = %q, want %q", contentStr, tt.expectedEnv)
				}
			case FormatJSON:
				if contentStr != tt.expectedJSON {
					t.Errorf("Write() json format = %q, want %q", contentStr, tt.expectedJSON)
				}
			}
		})
	}
}

func TestFileWriter_formatEnv(t *testing.T) {
	writer := NewFileWriter()

	tests := []struct {
		name     string
		vars     Variables
		expected string
	}{
		{
			name: "simple values",
			vars: Variables{
				{Key: "KEY1", Value: "value1"},
				{Key: "KEY2", Value: "value2"},
			},
			expected: "KEY1=value1\nKEY2=value2\n",
		},
		{
			name: "values with spaces",
			vars: Variables{
				{Key: "KEY1", Value: "value with spaces"},
				{Key: "KEY2", Value: "normal"},
			},
			expected: "KEY1=\"value with spaces\"\nKEY2=normal\n",
		},
		{
			name: "values with special characters",
			vars: Variables{
				{Key: "KEY1", Value: "value\nwith\nnewlines"},
				{Key: "KEY2", Value: "value\twith\ttabs"},
				{Key: "KEY3", Value: "value\"with\"quotes"},
			},
			expected: "KEY1=\"value\\nwith\\nnewlines\"\nKEY2=\"value\\twith\\ttabs\"\nKEY3=\"value\\\"with\\\"quotes\"\n",
		},
		{
			name:     "empty variables",
			vars:     Variables{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.formatEnv(tt.vars)
			if result != tt.expected {
				t.Errorf("formatEnv() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFileWriter_formatJSON(t *testing.T) {
	writer := NewFileWriter()

	tests := []struct {
		name     string
		vars     Variables
		expected string
	}{
		{
			name: "simple values",
			vars: Variables{
				{Key: "KEY1", Value: "value1"},
				{Key: "KEY2", Value: "value2"},
			},
			expected: `{"KEY1":"value1","KEY2":"value2"}`,
		},
		{
			name: "values with special characters",
			vars: Variables{
				{Key: "KEY1", Value: "value\"with\"quotes"},
				{Key: "KEY2", Value: "value\nwith\nnewlines"},
			},
			expected: `{"KEY1":"value\"with\"quotes","KEY2":"value\nwith\nnewlines"}`,
		},
		{
			name:     "empty variables",
			vars:     Variables{},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.formatJSON(tt.vars)
			if result != tt.expected {
				t.Errorf("formatJSON() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewFileLoader(t *testing.T) {
	loader := NewFileLoader()
	if loader == nil {
		t.Fatal("NewFileLoader() returned nil")
	}
}

func TestNewFileWriter(t *testing.T) {
	writer := NewFileWriter()
	if writer == nil {
		t.Fatal("NewFileWriter() returned nil")
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatEnv != "env" {
		t.Errorf("FormatEnv = %q, want %q", FormatEnv, "env")
	}

	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}

	if FormatYAML != "yaml" {
		t.Errorf("FormatYAML = %q, want %q", FormatYAML, "yaml")
	}
}

// Integration tests
func TestIntegration_LoadWriteRoundTrip(t *testing.T) {
	loader := NewFileLoader()
	writer := NewFileWriter()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test with simpler values that don't require complex escaping
	original := Variables{
		{Key: "KEY1", Value: "value1"},
		{Key: "KEY2", Value: "value with spaces"},
		{Key: "KEY3", Value: "simple_value"},
		{Key: "KEY4", Value: "another_value"},
	}

	// Only test env format since JSON format is output-only, not input
	t.Run("env_format", func(t *testing.T) {
		filename := filepath.Join(tempDir, "roundtrip_env.env")

		// Write
		err := writer.Write(filename, original, FormatEnv)
		if err != nil {
			t.Fatalf("Write() failed: %v", err)
		}

		// Read back
		loaded, err := loader.Load(context.Background(), filename)
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		// Compare
		if len(loaded) != len(original) {
			t.Errorf("Round-trip length mismatch: got %d, want %d", len(loaded), len(original))
		}

		loadedMap := loaded.ToMap()
		originalMap := original.ToMap()

		for k, originalValue := range originalMap {
			loadedValue, exists := loadedMap[k]
			if !exists {
				t.Errorf("Round-trip missing key: %s", k)
				continue
			}

			if loadedValue != originalValue {
				t.Errorf("Round-trip value mismatch for %s: got %q, want %q", k, loadedValue, originalValue)
			}
		}
	})

	// Test JSON output format separately (write-only test)
	t.Run("json_format_output", func(t *testing.T) {
		filename := filepath.Join(tempDir, "output_json.json")

		// Write in JSON format
		err := writer.Write(filename, original, FormatJSON)
		if err != nil {
			t.Fatalf("Write() JSON failed: %v", err)
		}

		// Verify the file contains valid JSON structure
		content, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read JSON file: %v", err)
		}

		jsonStr := string(content)
		if !strings.HasPrefix(jsonStr, "{") || !strings.HasSuffix(jsonStr, "}") {
			t.Errorf("JSON output doesn't have proper structure: %s", jsonStr)
		}

		// Verify it contains our keys
		for _, v := range original {
			if !strings.Contains(jsonStr, v.Key) {
				t.Errorf("JSON output missing key: %s", v.Key)
			}
		}
	})
}

func TestIntegration_EncryptionRoundTrip(t *testing.T) {
	loader := NewFileLoader()
	writer := NewFileWriter()
	encryptor := crypto.NewAESEncryptor()

	// Generate a test key
	key := make([]byte, crypto.KeySize)
	rand.Read(key)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	original := Variables{
		{Key: "SECRET1", Value: "secret_value_1"},
		{Key: "SECRET2", Value: "secret_value_2"},
		{Key: "PLAIN", Value: "plain_value"},
	}

	filename := filepath.Join(tempDir, "encrypted.env")

	// Encrypt sensitive values
	encrypted := make(Variables, len(original))
	copy(encrypted, original)

	for i, v := range encrypted {
		if strings.HasPrefix(v.Key, "SECRET") {
			encryptedValue, err := encryptor.Encrypt(v.Value, key)
			if err != nil {
				t.Fatal(err)
			}
			encrypted[i].Value = encryptedValue
		}
	}

	// Write encrypted values
	err = writer.Write(filename, encrypted, FormatEnv)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	// Load with decryption
	loaded, err := loader.LoadWithDecryption(context.Background(), filename, encryptor, key)
	if err != nil {
		t.Fatalf("LoadWithDecryption() failed: %v", err)
	}

	// Compare with original
	loadedMap := loaded.ToMap()
	originalMap := original.ToMap()

	for k, originalValue := range originalMap {
		loadedValue, exists := loadedMap[k]
		if !exists {
			t.Errorf("Encryption round-trip missing key: %s", k)
			continue
		}

		if loadedValue != originalValue {
			t.Errorf("Encryption round-trip value mismatch for %s: got %q, want %q", k, loadedValue, originalValue)
		}
	}
}

// Benchmark tests
func BenchmarkVariables_ToMap(b *testing.B) {
	vars := make(Variables, 100)
	for i := 0; i < 100; i++ {
		vars[i] = Variable{Key: "KEY" + string(rune(i)), Value: "value" + string(rune(i))}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vars.ToMap()
	}
}

func BenchmarkFileLoader_Load(b *testing.B) {
	loader := NewFileLoader()

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "bench_*.env")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Write test content
	content := strings.Repeat("KEY=value\n", 100)
	_, err = tempFile.WriteString(content)
	if err != nil {
		b.Fatal(err)
	}
	tempFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.Load(context.Background(), tempFile.Name())
		if err != nil {
			b.Fatal(err)
		}
	}
}
