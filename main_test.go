package main

import (
	"context"
	"crypto/rand"
	"os"
	"os/user"
	"testing"

	"github.com/almahoozi/envx/pkg/crypto"
	"github.com/almahoozi/envx/pkg/keystore"
)

func TestLoadEnv(t *testing.T) {
	// Create a temporary file for testing
	tempFile := createTempEnvFile(t, `KEY1=value1
KEY2=value2
# Comment
KEY3=value3`)
	defer removeTempFile(t, tempFile)

	vars, err := loadEnv(context.Background(), tempFile)
	if err != nil {
		t.Errorf("loadEnv() unexpected error: %v", err)
	}

	expected := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "value3",
	}

	varMap := vars.ToMap()
	if len(varMap) != len(expected) {
		t.Errorf("loadEnv() returned %d variables, want %d", len(varMap), len(expected))
	}

	for k, v := range expected {
		if varMap[k] != v {
			t.Errorf("loadEnv() %s = %q, want %q", k, varMap[k], v)
		}
	}
}

func TestLoadDecryptedEnv(t *testing.T) {
	encryptor := crypto.NewAESEncryptor()
	key := generateTestKey(t)

	// Create encrypted values
	secret1, err := encryptor.Encrypt("secret_value_1", key)
	if err != nil {
		t.Fatal(err)
	}

	secret2, err := encryptor.Encrypt("secret_value_2", key)
	if err != nil {
		t.Fatal(err)
	}

	content := "SECRET1=" + secret1 + "\n" +
		"SECRET2=" + secret2 + "\n" +
		"PLAIN=plain_value"

	tempFile := createTempEnvFile(t, content)
	defer removeTempFile(t, tempFile)

	vars, err := loadDecryptedEnv(context.Background(), tempFile, encryptor, key)
	if err != nil {
		t.Errorf("loadDecryptedEnv() unexpected error: %v", err)
	}

	expected := map[string]string{
		"SECRET1": "secret_value_1",
		"SECRET2": "secret_value_2",
		"PLAIN":   "plain_value",
	}

	varMap := vars.ToMap()
	for k, v := range expected {
		if varMap[k] != v {
			t.Errorf("loadDecryptedEnv() %s = %q, want %q", k, varMap[k], v)
		}
	}
}

func TestLoadKey(t *testing.T) {
	// Setup test keystore to avoid interfering with production keychain
	setupTestKeystore(t)
	defer teardownTestKeystore(t)

	// Get current user
	currentUser, err := user.Current()
	if err != nil {
		t.Skip("Cannot get current user, skipping loadKey test")
	}

	// Use test-specific keychain config to avoid interfering with production
	testConfig := &keystore.Config{
		App:     "envx.test",
		Service: "com.almahoozi.envx.test",
	}

	// Test loadKeyWithConfig using test config
	key, err := loadKeyWithConfig(testConfig)
	if err != nil {
		t.Errorf("loadKeyWithConfig() unexpected error: %v", err)
	}

	if len(key) != crypto.KeySize {
		t.Errorf("loadKeyWithConfig() returned key of size %d, want %d", len(key), crypto.KeySize)
	}

	// Test that calling loadKeyWithConfig again returns the same key
	key2, err := loadKeyWithConfig(testConfig)
	if err != nil {
		t.Errorf("loadKeyWithConfig() second call unexpected error: %v", err)
	}

	if len(key2) != crypto.KeySize {
		t.Errorf("loadKeyWithConfig() second call returned key of size %d, want %d", len(key2), crypto.KeySize)
	}

	// Keys should be the same
	for i, b := range key {
		if key2[i] != b {
			t.Errorf("loadKeyWithConfig() returned different key on second call")
			break
		}
	}

	t.Logf("loadKey() test completed for user: %s (using test keychain item)", currentUser.Username)
}

// Helper functions for tests

func createTempEnvFile(t *testing.T, content string) string {
	t.Helper()

	file, err := os.CreateTemp("", "envx_test_*.env")
	if err != nil {
		t.Fatal(err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		t.Fatal(err)
	}

	file.Close()
	return file.Name()
}

func removeTempFile(t *testing.T, filename string) {
	t.Helper()

	err := os.Remove(filename)
	if err != nil {
		t.Errorf("Failed to remove temp file %s: %v", filename, err)
	}
}

func generateTestKey(t *testing.T) []byte {
	t.Helper()

	key := make([]byte, crypto.KeySize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatal(err)
	}

	return key
}

// setupTestKeystore configures the global keystore to use test-specific keychain items
func setupTestKeystore(t *testing.T) {
	t.Helper()

	// Use mock keystore for all tests to avoid platform dependencies
	testKeystore = keystore.NewMockKeyStore()

	testKeystoreConfig = &keystore.Config{
		App:     "envx.test",
		Service: "com.almahoozi.envx.test",
	}
}

// teardownTestKeystore resets the global keystore configuration
func teardownTestKeystore(t *testing.T) {
	t.Helper()

	testKeystore = nil
	testKeystoreConfig = nil
}
