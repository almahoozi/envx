package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFmtOpts_Format(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		json     bool
		yaml     bool
		yml      bool
		expected Format
		wantErr  bool
	}{
		{
			name:     "env format",
			format:   "env",
			expected: FormatEnv,
			wantErr:  false,
		},
		{
			name:     "json format",
			format:   "json",
			expected: FormatJSON,
			wantErr:  false,
		},
		{
			name:     "yaml format",
			format:   "yaml",
			expected: FormatYAML,
			wantErr:  false,
		},
		{
			name:     "json flag",
			json:     true,
			expected: FormatJSON,
			wantErr:  false,
		},
		{
			name:     "yaml flag",
			yaml:     true,
			expected: FormatYAML,
			wantErr:  false,
		},
		{
			name:     "yml flag",
			yml:      true,
			expected: FormatYAML,
			wantErr:  false,
		},
		{
			name:     "default format",
			expected: FormatEnv,
			wantErr:  false,
		},
		{
			name:    "conflicting format and json",
			format:  "env",
			json:    true,
			wantErr: true,
		},
		{
			name:    "conflicting json and yaml",
			json:    true,
			yaml:    true,
			wantErr: true,
		},
		{
			name:    "conflicting yaml and yml",
			yaml:    true,
			yml:     true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &fmtOpts{
				format: tt.format,
				json:   tt.json,
				yaml:   tt.yaml,
				yml:    tt.yml,
			}

			result, err := opts.Format()

			if tt.wantErr {
				if err == nil {
					t.Error("Format() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Format() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Format() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetVCmdFn(t *testing.T) {
	// Setup test keystore to avoid interfering with production keychain
	setupTestKeystore(t)
	defer teardownTestKeystore(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test environment file
	envFile := filepath.Join(tempDir, ".env")
	content := `KEY1=value1
KEY2=value2
KEY3=value3`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		opts      getVOpts
		args      []string
		wantErr   bool
		wantCount int
	}{
		{
			name: "get all values",
			opts: getVOpts{
				File:      envFile,
				Name:      "",
				KeyStore:  "mock",
				KeyName:   "",
				Separator: "\n",
			},
			args:      []string{},
			wantErr:   false,
			wantCount: 3,
		},
		{
			name: "get specific values",
			opts: getVOpts{
				File:      envFile,
				Name:      "",
				KeyStore:  "mock",
				KeyName:   "",
				Separator: ",",
			},
			args:      []string{"KEY1", "KEY3"},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "get non-existent key",
			opts: getVOpts{
				File:      envFile,
				Name:      "",
				KeyStore:  "mock",
				KeyName:   "",
				Separator: "\n",
			},
			args:    []string{"NONEXISTENT"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output by redirecting stdout
			// Note: In a real test environment, you might use a more sophisticated
			// method to capture output
			err := getVCmdFn(context.Background(), tt.opts, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("getVCmdFn() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("getVCmdFn() unexpected error: %v", err)
			}
		})
	}
}

func TestGetCmdFn(t *testing.T) {
	// Setup test keystore to avoid interfering with production keychain
	setupTestKeystore(t)
	defer teardownTestKeystore(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test environment file
	envFile := filepath.Join(tempDir, ".env")
	content := `KEY1=value1
KEY2=value2
KEY3=value3`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    getOpts
		args    []string
		wantErr bool
	}{
		{
			name: "get all env format",
			opts: getOpts{
				File:       envFile,
				Name:       "",
				KeyStore:   "mock",
				KeyName:    "",
				FmtOpts:    &fmtOpts{format: "env"},
				ValuesOnly: false,
			},
			args:    []string{},
			wantErr: false,
		},
		{
			name: "get specific keys",
			opts: getOpts{
				File:       envFile,
				Name:       "",
				KeyStore:   "mock",
				KeyName:    "",
				FmtOpts:    &fmtOpts{json: true},
				ValuesOnly: false,
			},
			args:    []string{"KEY1", "KEY2"},
			wantErr: false,
		},
		{
			name: "values only",
			opts: getOpts{
				File:       envFile,
				Name:       "",
				KeyStore:   "mock",
				KeyName:    "",
				FmtOpts:    &fmtOpts{},
				ValuesOnly: true,
			},
			args:    []string{},
			wantErr: false,
		},
		{
			name: "yaml format - unsupported",
			opts: getOpts{
				File:       envFile,
				Name:       "",
				KeyStore:   "mock",
				KeyName:    "",
				FmtOpts:    &fmtOpts{yaml: true},
				ValuesOnly: false,
			},
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := getCmdFn(context.Background(), tt.opts, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("getCmdFn() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("getCmdFn() unexpected error: %v", err)
			}
		})
	}
}

func TestSetCmdFn(t *testing.T) {
	// Setup test keystore to avoid interfering with production keychain
	setupTestKeystore(t)
	defer teardownTestKeystore(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test environment file
	envFile := filepath.Join(tempDir, ".env")
	content := `KEY1=value1
KEY2=value2`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    setOpts
		args    []string
		wantErr bool
	}{
		{
			name: "set existing key",
			opts: setOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
				FmtOpts:  &fmtOpts{format: "env"},
				print:    true,
			},
			args:    []string{"KEY1=new_value1"},
			wantErr: false,
		},
		{
			name: "set new key",
			opts: setOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
				FmtOpts:  &fmtOpts{json: true},
				print:    true,
			},
			args:    []string{"NEW_KEY=new_value"},
			wantErr: false,
		},
		{
			name: "key only format - would prompt for input",
			opts: setOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
				FmtOpts:  &fmtOpts{},
				print:    true,
			},
			args:    []string{"KEY_ONLY"},
			wantErr: true, // Will fail in test because it tries to read from stdin
		},
		{
			name: "yaml format - unsupported",
			opts: setOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
				FmtOpts:  &fmtOpts{yaml: true},
				print:    true,
			},
			args:    []string{"KEY=value"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setCmdFn(context.Background(), tt.opts, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("setCmdFn() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("setCmdFn() unexpected error: %v", err)
			}
		})
	}
}

func TestAddCmdFn(t *testing.T) {
	// Setup test keystore to avoid interfering with production keychain
	setupTestKeystore(t)
	defer teardownTestKeystore(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test environment file
	envFile := filepath.Join(tempDir, ".env")
	content := `KEY1=value1
KEY2=value2`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    addOpts
		args    []string
		wantErr bool
	}{
		{
			name: "add new key",
			opts: addOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
				FmtOpts:  &fmtOpts{format: "env"},
				print:    true,
			},
			args:    []string{"NEW_KEY=new_value"},
			wantErr: false,
		},
		{
			name: "add existing key - should fail",
			opts: addOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
				FmtOpts:  &fmtOpts{},
				print:    true,
			},
			args:    []string{"KEY1=new_value"},
			wantErr: true,
		},
		{
			name: "key only format - would prompt for input",
			opts: addOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
				FmtOpts:  &fmtOpts{},
				print:    true,
			},
			args:    []string{"KEY_ONLY"},
			wantErr: true, // Will fail in test because it tries to read from stdin
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := addCmdFn(context.Background(), tt.opts, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("addCmdFn() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("addCmdFn() unexpected error: %v", err)
			}
		})
	}
}

func TestEncryptCmd(t *testing.T) {
	// Setup test keystore to avoid interfering with production keychain
	setupTestKeystore(t)
	defer teardownTestKeystore(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test environment file
	envFile := filepath.Join(tempDir, ".env")
	content := `SECRET1=secret_value_1
SECRET2=secret_value_2
PLAIN=plain_value`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    encryptOpts
		args    []string
		wantErr bool
	}{
		{
			name: "encrypt all values",
			opts: encryptOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
				FmtOpts:  &fmtOpts{format: "env"},
				Write:    false,
			},
			args:    []string{},
			wantErr: false,
		},
		{
			name: "encrypt specific values",
			opts: encryptOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
				FmtOpts:  &fmtOpts{json: true},
				Write:    false,
			},
			args:    []string{"SECRET1", "SECRET2"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := encryptCmd(context.Background(), tt.opts, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("encryptCmd() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("encryptCmd() unexpected error: %v", err)
			}
		})
	}
}

func TestDecryptCmd(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Note: We skip this test because it requires the actual keychain
	// which uses different keys than what we generate in the test
	t.Skip("Skipping decrypt test as it requires keychain integration")
}

func TestRun(t *testing.T) {
	// Setup test keystore to avoid interfering with production keychain
	setupTestKeystore(t)
	defer teardownTestKeystore(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test environment file
	envFile := filepath.Join(tempDir, ".env")
	content := `TEST_VAR=test_value
ANOTHER_VAR=another_value`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		opts    runOpts
		args    []string
		wantErr bool
	}{
		{
			name: "missing executable",
			opts: runOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
			},
			args:    []string{},
			wantErr: true,
		},
		{
			name: "non-existent executable",
			opts: runOpts{
				File:     envFile,
				Name:     "",
				KeyStore: "mock",
				KeyName:  "",
			},
			args:    []string{"non_existent_executable_12345"},
			wantErr: true,
		},
		// Note: We can't easily test successful execution without creating
		// a real executable, which would be complex in a unit test
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(context.Background(), tt.opts, tt.args...)

			if tt.wantErr {
				if err == nil {
					t.Error("run() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("run() unexpected error: %v", err)
			}
		})
	}
}

// Integration test for command execution flow
func TestCommandExecutionFlow(t *testing.T) {
	// Setup test keystore to avoid interfering with production keychain
	setupTestKeystore(t)
	defer teardownTestKeystore(t)

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "envx_cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	envFile := filepath.Join(tempDir, ".env")

	// Test the full flow: add -> encrypt -> decrypt -> get
	t.Run("full_workflow", func(t *testing.T) {
		// Step 1: Add a new variable
		addOpts := addOpts{
			File:     envFile,
			Name:     "",
			KeyStore: "mock",
			KeyName:  "",
			FmtOpts:  &fmtOpts{format: "env"},
			print:    false,
		}

		err := addCmdFn(context.Background(), addOpts, "SECRET=my_secret_value")
		if err != nil {
			t.Fatalf("addCmdFn() failed: %v", err)
		}

		// Verify the file was created and contains the variable
		content, err := os.ReadFile(envFile)
		if err != nil {
			t.Fatalf("Failed to read env file: %v", err)
		}

		if !strings.Contains(string(content), "SECRET=") {
			t.Errorf("Environment file doesn't contain expected variable")
		}

		// Step 2: Encrypt the variable
		encryptOpts := encryptOpts{
			File:     envFile,
			Name:     "",
			KeyStore: "mock",
			KeyName:  "",
			FmtOpts:  &fmtOpts{format: "env"},
			Write:    true,
		}

		err = encryptCmd(context.Background(), encryptOpts, "SECRET")
		if err != nil {
			t.Fatalf("encryptCmd() failed: %v", err)
		}

		// Verify the file now contains encrypted data
		content, err = os.ReadFile(envFile)
		if err != nil {
			t.Fatalf("Failed to read env file after encryption: %v", err)
		}

		if strings.Contains(string(content), "my_secret_value") {
			t.Errorf("Environment file still contains plaintext after encryption")
		}

		// Step 3: Decrypt and verify
		decryptOpts := decryptOpts{
			File:     envFile,
			Name:     "",
			KeyStore: "mock",
			KeyName:  "",
			FmtOpts:  &fmtOpts{format: "env"},
			Write:    false,
		}

		err = decryptCmd(context.Background(), decryptOpts)
		if err != nil {
			t.Fatalf("decryptCmd() failed: %v", err)
		}

		// Step 4: Get the decrypted value
		getOpts := getOpts{
			File:       envFile,
			Name:       "",
			KeyStore:   "mock",
			KeyName:    "",
			FmtOpts:    &fmtOpts{format: "env"},
			ValuesOnly: false,
		}

		err = getCmdFn(context.Background(), getOpts, "SECRET")
		if err != nil {
			t.Fatalf("getCmdFn() failed: %v", err)
		}
	})
}

func TestParseKeyValueArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "key=value format",
			args:     []string{"KEY1=value1", "KEY2=value2"},
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
			wantErr:  false,
		},
		{
			name:     "empty key with equals",
			args:     []string{"=value"},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "empty key",
			args:     []string{""},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "key with spaces",
			args:     []string{"  KEY1  =  value1  "},
			expected: map[string]string{"KEY1": "value1"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For tests that don't involve prompting (key=value format)
			if tt.wantErr || containsEquals(tt.args) {
				result, err := parseKeyValueArgs(tt.args)

				if tt.wantErr {
					if err == nil {
						t.Errorf("parseKeyValueArgs() expected error but got none")
					}
					return
				}

				if err != nil {
					t.Errorf("parseKeyValueArgs() unexpected error: %v", err)
					return
				}

				if len(result) != len(tt.expected) {
					t.Errorf("parseKeyValueArgs() got %d keys, want %d", len(result), len(tt.expected))
					return
				}

				for k, v := range tt.expected {
					if result[k] != v {
						t.Errorf("parseKeyValueArgs() key %s = %q, want %q", k, result[k], v)
					}
				}
			}
		})
	}
}

func containsEquals(args []string) bool {
	for _, arg := range args {
		if !strings.Contains(arg, "=") {
			return false
		}
	}
	return true
}
