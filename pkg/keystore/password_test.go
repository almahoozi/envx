package keystore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/almahoozi/envx/pkg/crypto"
)

func TestNewPasswordKeyStore(t *testing.T) {
	// Test with nil config
	store := NewPasswordKeyStore(nil)
	passwordStore := store.(*PasswordKeyStore)

	if passwordStore.iterations != DefaultIterations {
		t.Errorf("Expected iterations to be %d, got %d", DefaultIterations, passwordStore.iterations)
	}

	// Test with custom config
	config := &PasswordKeyStoreConfig{
		Iterations: 50000,
		PromptFunc: func(prompt string) (string, error) { return "test", nil },
	}
	store = NewPasswordKeyStore(config)
	passwordStore = store.(*PasswordKeyStore)

	if passwordStore.iterations != 50000 {
		t.Errorf("Expected iterations to be 50000, got %d", passwordStore.iterations)
	}
}

func TestPasswordKeyStore_CreateKey(t *testing.T) {
	// Create a temporary directory for salt files
	tempDir := t.TempDir()

	// Mock prompt function
	promptCalls := 0
	mockPrompt := func(prompt string) (string, error) {
		promptCalls++
		return "testpassword", nil
	}

	config := &PasswordKeyStoreConfig{
		Iterations: 1000, // Use fewer iterations for faster tests
		PromptFunc: mockPrompt,
	}

	store := NewPasswordKeyStore(config).(*PasswordKeyStore)

	// Override salt directory for testing
	originalGetSaltDir := getSaltDir
	defer func() { getSaltDir = originalGetSaltDir }()
	getSaltDir = func() string { return tempDir }

	account := "testaccount"
	key, err := store.CreateKey(account)

	if err != nil {
		t.Fatalf("CreateKey failed: %v", err)
	}

	if len(key) != crypto.KeySize {
		t.Errorf("Expected key size %d, got %d", crypto.KeySize, len(key))
	}

	// Should have prompted twice (create and confirm)
	if promptCalls != 2 {
		t.Errorf("Expected 2 prompt calls, got %d", promptCalls)
	}

	// Check that salt file was created
	saltFile := filepath.Join(tempDir, account+".salt")
	if _, err := os.Stat(saltFile); os.IsNotExist(err) {
		t.Error("Salt file was not created")
	}
}

func TestPasswordKeyStore_CreateKey_PasswordMismatch(t *testing.T) {
	tempDir := t.TempDir()

	// Mock prompt function that returns different passwords
	promptCalls := 0
	mockPrompt := func(prompt string) (string, error) {
		promptCalls++
		if promptCalls == 1 {
			return "password1", nil
		}
		return "password2", nil
	}

	config := &PasswordKeyStoreConfig{
		Iterations: 1000,
		PromptFunc: mockPrompt,
	}

	store := NewPasswordKeyStore(config).(*PasswordKeyStore)

	// Override salt directory for testing
	originalGetSaltDir := getSaltDir
	defer func() { getSaltDir = originalGetSaltDir }()
	getSaltDir = func() string { return tempDir }

	_, err := store.CreateKey("testaccount")

	if err == nil {
		t.Error("Expected error for password mismatch, got nil")
	}

	if err.Error() != "passwords do not match" {
		t.Errorf("Expected 'passwords do not match' error, got: %v", err)
	}
}

func TestPasswordKeyStore_GetKey(t *testing.T) {
	tempDir := t.TempDir()

	// Mock prompt function
	mockPrompt := func(prompt string) (string, error) {
		return "testpassword", nil
	}

	config := &PasswordKeyStoreConfig{
		Iterations: 1000,
		PromptFunc: mockPrompt,
	}

	store := NewPasswordKeyStore(config).(*PasswordKeyStore)

	// Override salt directory for testing
	originalGetSaltDir := getSaltDir
	defer func() { getSaltDir = originalGetSaltDir }()
	getSaltDir = func() string { return tempDir }

	account := "testaccount"

	// First create a key to establish the salt
	originalKey, err := store.CreateKey(account)
	if err != nil {
		t.Fatalf("CreateKey failed: %v", err)
	}

	// Now get the key again - should be the same
	retrievedKey, err := store.GetKey(account)
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}

	// Keys should be identical since same password and salt
	if len(retrievedKey) != len(originalKey) {
		t.Errorf("Key lengths don't match: %d vs %d", len(retrievedKey), len(originalKey))
	}

	for i := range originalKey {
		if originalKey[i] != retrievedKey[i] {
			t.Error("Keys don't match")
			break
		}
	}
}

func TestPasswordKeyStore_LoadOrCreateKey(t *testing.T) {
	tempDir := t.TempDir()

	mockPrompt := func(prompt string) (string, error) {
		return "testpassword", nil
	}

	config := &PasswordKeyStoreConfig{
		Iterations: 1000,
		PromptFunc: mockPrompt,
	}

	store := NewPasswordKeyStore(config).(*PasswordKeyStore)

	// Override salt directory for testing
	originalGetSaltDir := getSaltDir
	defer func() { getSaltDir = originalGetSaltDir }()
	getSaltDir = func() string { return tempDir }

	account := "testaccount"

	// First call should create the key
	key1, err := store.LoadOrCreateKey(account)
	if err != nil {
		t.Fatalf("LoadOrCreateKey failed: %v", err)
	}

	// Second call should load the existing key
	key2, err := store.LoadOrCreateKey(account)
	if err != nil {
		t.Fatalf("LoadOrCreateKey failed on second call: %v", err)
	}

	// Keys should be identical
	if len(key1) != len(key2) {
		t.Errorf("Key lengths don't match: %d vs %d", len(key1), len(key2))
	}

	for i := range key1 {
		if key1[i] != key2[i] {
			t.Error("Keys don't match")
			break
		}
	}
}

func TestPasswordKeyStore_SetKey(t *testing.T) {
	store := NewPasswordKeyStore(nil)

	// SetKey should not be supported
	err := store.SetKey("testaccount", make([]byte, crypto.KeySize))
	if err == nil {
		t.Error("Expected error for SetKey, got nil")
	}

	expectedMsg := "SetKey not supported for password-based keystore"
	if err.Error() != expectedMsg+" - keys are derived from passwords" {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestPasswordKeyStore_DifferentPasswords(t *testing.T) {
	tempDir := t.TempDir()

	// Create two stores with different passwords
	mockPrompt1 := func(prompt string) (string, error) {
		return "password1", nil
	}

	mockPrompt2 := func(prompt string) (string, error) {
		return "password2", nil
	}

	config1 := &PasswordKeyStoreConfig{
		Iterations: 1000,
		PromptFunc: mockPrompt1,
	}

	config2 := &PasswordKeyStoreConfig{
		Iterations: 1000,
		PromptFunc: mockPrompt2,
	}

	store1 := NewPasswordKeyStore(config1).(*PasswordKeyStore)
	store2 := NewPasswordKeyStore(config2).(*PasswordKeyStore)

	// Override salt directory for testing
	originalGetSaltDir := getSaltDir
	defer func() { getSaltDir = originalGetSaltDir }()
	getSaltDir = func() string { return tempDir }

	account := "testaccount"

	// Create key with first password
	key1, err := store1.CreateKey(account)
	if err != nil {
		t.Fatalf("CreateKey failed: %v", err)
	}

	// Get key with second password (same salt, different password)
	key2, err := store2.GetKey(account)
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}

	// Keys should be different
	same := true
	for i := range key1 {
		if key1[i] != key2[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("Keys should be different with different passwords")
	}
}

func TestPasswordKeyStore_Constants(t *testing.T) {
	if DefaultIterations < 10000 {
		t.Errorf("DefaultIterations should be at least 10000 for security, got %d", DefaultIterations)
	}

	if SaltSize < 16 {
		t.Errorf("SaltSize should be at least 16 bytes for security, got %d", SaltSize)
	}
}

func TestPasswordKeyStore_NonInteractivePassword(t *testing.T) {
	tempDir := t.TempDir()

	// Test with password provided in config
	config := &PasswordKeyStoreConfig{
		Iterations: 1000,
		Password:   "testpassword123",
	}

	store := NewPasswordKeyStore(config).(*PasswordKeyStore)

	// Override salt directory for testing
	originalGetSaltDir := getSaltDir
	defer func() { getSaltDir = originalGetSaltDir }()
	getSaltDir = func() string { return tempDir }

	account := "testaccount"

	// Create key should work without prompting
	key1, err := store.CreateKey(account)
	if err != nil {
		t.Fatalf("CreateKey failed: %v", err)
	}

	if len(key1) != crypto.KeySize {
		t.Errorf("Expected key size %d, got %d", crypto.KeySize, len(key1))
	}

	// Get key should work without prompting
	key2, err := store.GetKey(account)
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}

	// Keys should be identical
	if len(key1) != len(key2) {
		t.Errorf("Key lengths don't match: %d vs %d", len(key1), len(key2))
	}

	for i := range key1 {
		if key1[i] != key2[i] {
			t.Error("Keys don't match")
			break
		}
	}
}

func TestPasswordKeyStore_EnvironmentVariable(t *testing.T) {
	tempDir := t.TempDir()

	// Set environment variable
	originalEnv := os.Getenv("ENVX_PASSWORD")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENVX_PASSWORD")
		} else {
			os.Setenv("ENVX_PASSWORD", originalEnv)
		}
	}()

	os.Setenv("ENVX_PASSWORD", "envvarpassword")

	// Create store without password in config
	config := &PasswordKeyStoreConfig{
		Iterations: 1000,
	}

	store := NewPasswordKeyStore(config).(*PasswordKeyStore)

	// Override salt directory for testing
	originalGetSaltDir := getSaltDir
	defer func() { getSaltDir = originalGetSaltDir }()
	getSaltDir = func() string { return tempDir }

	account := "testaccount"

	// Should use environment variable password
	key, err := store.CreateKey(account)
	if err != nil {
		t.Fatalf("CreateKey failed: %v", err)
	}

	if len(key) != crypto.KeySize {
		t.Errorf("Expected key size %d, got %d", crypto.KeySize, len(key))
	}
}

func TestPasswordKeyStore_PasswordPriority(t *testing.T) {
	tempDir := t.TempDir()

	// Set environment variable
	originalEnv := os.Getenv("ENVX_PASSWORD")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("ENVX_PASSWORD")
		} else {
			os.Setenv("ENVX_PASSWORD", originalEnv)
		}
	}()

	os.Setenv("ENVX_PASSWORD", "envvarpassword")

	// Create store with password in config
	config := &PasswordKeyStoreConfig{
		Iterations: 1000,
		Password:   "configpassword",
	}

	store := NewPasswordKeyStore(config).(*PasswordKeyStore)

	// Override salt directory for testing
	originalGetSaltDir := getSaltDir
	defer func() { getSaltDir = originalGetSaltDir }()
	getSaltDir = func() string { return tempDir }

	account := "testaccount"

	// Create key with env var (should take priority)
	key1, err := store.CreateKey(account)
	if err != nil {
		t.Fatalf("CreateKey failed: %v", err)
	}

	// Unset env var
	os.Unsetenv("ENVX_PASSWORD")

	// Get key with config password (should work since salt exists)
	key2, err := store.GetKey(account)
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}

	// Keys should be different because different passwords were used
	same := true
	for i := range key1 {
		if key1[i] != key2[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("Keys should be different when using different passwords")
	}
}
