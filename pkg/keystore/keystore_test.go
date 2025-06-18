package keystore

import (
	"crypto/rand"
	"errors"
	"testing"

	"github.com/almahoozi/envx/pkg/crypto"
)

// Mock implementation for testing
type mockKeyStore struct {
	keys   map[string][]byte
	errors map[string]error // map of account to error to return
}

func newMockKeyStore() *mockKeyStore {
	return &mockKeyStore{
		keys:   make(map[string][]byte),
		errors: make(map[string]error),
	}
}

func (m *mockKeyStore) GetKey(account string) ([]byte, error) {
	if err, exists := m.errors[account]; exists {
		return nil, err
	}

	key, exists := m.keys[account]
	if !exists {
		return nil, errors.New("key not found")
	}

	return key, nil
}

func (m *mockKeyStore) SetKey(account string, key []byte) error {
	if err, exists := m.errors[account]; exists {
		return err
	}

	if len(key) != crypto.KeySize {
		return errors.New("invalid key size")
	}

	m.keys[account] = make([]byte, len(key))
	copy(m.keys[account], key)
	return nil
}

func (m *mockKeyStore) CreateKey(account string) ([]byte, error) {
	if err, exists := m.errors[account]; exists {
		return nil, err
	}

	key := make([]byte, crypto.KeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	err = m.SetKey(account, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (m *mockKeyStore) LoadOrCreateKey(account string) ([]byte, error) {
	key, err := m.GetKey(account)
	if err == nil && len(key) == crypto.KeySize {
		return key, nil
	}

	return m.CreateKey(account)
}

// Set an error to be returned for a specific account
func (m *mockKeyStore) setError(account string, err error) {
	m.errors[account] = err
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.App != "envx" {
		t.Errorf("DefaultConfig().App = %q, want %q", config.App, "envx")
	}

	if config.Service != "com.almahoozi.envx" {
		t.Errorf("DefaultConfig().Service = %q, want %q", config.Service, "com.almahoozi.envx")
	}
}

func TestNewMacOSKeyStore(t *testing.T) {
	// Test with nil config
	store := NewMacOSKeyStore(nil)
	if store == nil {
		t.Fatal("NewMacOSKeyStore(nil) returned nil")
	}

	// Test with custom config
	config := &Config{
		App:     "test-app",
		Service: "test.service",
	}
	store = NewMacOSKeyStore(config)
	if store == nil {
		t.Fatal("NewMacOSKeyStore(config) returned nil")
	}
}

func TestMockKeyStore_GetKey(t *testing.T) {
	store := newMockKeyStore()

	// Test key not found
	_, err := store.GetKey("nonexistent")
	if err == nil {
		t.Error("GetKey() expected error for nonexistent key")
	}

	// Test existing key
	testKey := make([]byte, crypto.KeySize)
	rand.Read(testKey)
	store.keys["test"] = testKey

	retrievedKey, err := store.GetKey("test")
	if err != nil {
		t.Errorf("GetKey() unexpected error: %v", err)
	}

	if len(retrievedKey) != len(testKey) {
		t.Errorf("GetKey() returned key of wrong length: got %d, want %d", len(retrievedKey), len(testKey))
	}

	// Verify key contents
	for i, b := range testKey {
		if retrievedKey[i] != b {
			t.Errorf("GetKey() returned different key at position %d", i)
			break
		}
	}

	// Test error condition
	store.setError("error-account", errors.New("test error"))
	_, err = store.GetKey("error-account")
	if err == nil {
		t.Error("GetKey() expected error but got none")
	}
}

func TestMockKeyStore_SetKey(t *testing.T) {
	store := newMockKeyStore()

	tests := []struct {
		name    string
		account string
		key     []byte
		wantErr bool
	}{
		{
			name:    "valid key",
			account: "test1",
			key:     make([]byte, crypto.KeySize),
			wantErr: false,
		},
		{
			name:    "invalid key size - too short",
			account: "test2",
			key:     make([]byte, 16),
			wantErr: true,
		},
		{
			name:    "invalid key size - too long",
			account: "test3",
			key:     make([]byte, 64),
			wantErr: true,
		},
		{
			name:    "nil key",
			account: "test4",
			key:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key != nil && len(tt.key) == crypto.KeySize {
				rand.Read(tt.key) // Fill with random data
			}

			err := store.SetKey(tt.account, tt.key)

			if tt.wantErr {
				if err == nil {
					t.Error("SetKey() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("SetKey() unexpected error: %v", err)
				return
			}

			// Verify key was stored correctly
			retrievedKey, err := store.GetKey(tt.account)
			if err != nil {
				t.Errorf("Failed to retrieve stored key: %v", err)
				return
			}

			if len(retrievedKey) != len(tt.key) {
				t.Errorf("Retrieved key has wrong length: got %d, want %d", len(retrievedKey), len(tt.key))
			}

			for i, b := range tt.key {
				if retrievedKey[i] != b {
					t.Errorf("Retrieved key differs at position %d", i)
					break
				}
			}
		})
	}

	// Test error condition
	store.setError("error-account", errors.New("test error"))
	err := store.SetKey("error-account", make([]byte, crypto.KeySize))
	if err == nil {
		t.Error("SetKey() expected error but got none")
	}
}

func TestMockKeyStore_CreateKey(t *testing.T) {
	store := newMockKeyStore()

	// Test successful key creation
	key, err := store.CreateKey("test")
	if err != nil {
		t.Errorf("CreateKey() unexpected error: %v", err)
	}

	if len(key) != crypto.KeySize {
		t.Errorf("CreateKey() returned key of wrong size: got %d, want %d", len(key), crypto.KeySize)
	}

	// Verify key was stored
	retrievedKey, err := store.GetKey("test")
	if err != nil {
		t.Errorf("Failed to retrieve created key: %v", err)
	}

	for i, b := range key {
		if retrievedKey[i] != b {
			t.Errorf("Created key differs from stored key at position %d", i)
			break
		}
	}

	// Test that multiple keys are different
	key2, err := store.CreateKey("test2")
	if err != nil {
		t.Errorf("CreateKey() second call unexpected error: %v", err)
	}

	// Keys should be different
	same := true
	for i := 0; i < len(key) && i < len(key2); i++ {
		if key[i] != key2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("CreateKey() generated identical keys")
	}

	// Test error condition
	store.setError("error-account", errors.New("test error"))
	_, err = store.CreateKey("error-account")
	if err == nil {
		t.Error("CreateKey() expected error but got none")
	}
}

func TestMockKeyStore_LoadOrCreateKey(t *testing.T) {
	store := newMockKeyStore()

	// Test creating new key when none exists
	key1, err := store.LoadOrCreateKey("new-account")
	if err != nil {
		t.Errorf("LoadOrCreateKey() unexpected error: %v", err)
	}

	if len(key1) != crypto.KeySize {
		t.Errorf("LoadOrCreateKey() returned key of wrong size: got %d, want %d", len(key1), crypto.KeySize)
	}

	// Test loading existing key
	key2, err := store.LoadOrCreateKey("new-account")
	if err != nil {
		t.Errorf("LoadOrCreateKey() unexpected error on second call: %v", err)
	}

	// Should return the same key
	for i, b := range key1 {
		if key2[i] != b {
			t.Errorf("LoadOrCreateKey() returned different key on second call at position %d", i)
			break
		}
	}

	// Test with corrupted key (wrong size)
	store.keys["corrupted"] = make([]byte, 16) // wrong size
	key3, err := store.LoadOrCreateKey("corrupted")
	if err != nil {
		t.Errorf("LoadOrCreateKey() unexpected error with corrupted key: %v", err)
	}

	if len(key3) != crypto.KeySize {
		t.Errorf("LoadOrCreateKey() didn't recreate key with correct size: got %d, want %d", len(key3), crypto.KeySize)
	}

	// Test error condition
	store.setError("error-account", errors.New("test error"))
	_, err = store.LoadOrCreateKey("error-account")
	if err == nil {
		t.Error("LoadOrCreateKey() expected error but got none")
	}
}

func TestKeyStoreInterface(t *testing.T) {
	// Test that mockKeyStore implements KeyStore interface
	var _ KeyStore = &mockKeyStore{}

	// Test that macOSKeyStore implements KeyStore interface
	var _ KeyStore = NewMacOSKeyStore(nil)
}

// Integration test for keystore operations
func TestKeyStoreOperations(t *testing.T) {
	stores := map[string]KeyStore{
		"mock": newMockKeyStore(),
		// Note: We don't include macOS keystore in automated tests
		// as it would interfere with the actual keychain
	}

	for name, store := range stores {
		t.Run(name, func(t *testing.T) {
			account := "test-integration-" + name

			// Test LoadOrCreateKey creates a new key
			key1, err := store.LoadOrCreateKey(account)
			if err != nil {
				t.Fatalf("LoadOrCreateKey() failed: %v", err)
			}

			if len(key1) != crypto.KeySize {
				t.Errorf("Key has wrong size: got %d, want %d", len(key1), crypto.KeySize)
			}

			// Test LoadOrCreateKey returns the same key
			key2, err := store.LoadOrCreateKey(account)
			if err != nil {
				t.Fatalf("LoadOrCreateKey() second call failed: %v", err)
			}

			for i, b := range key1 {
				if key2[i] != b {
					t.Errorf("LoadOrCreateKey() returned different key on second call")
					break
				}
			}

			// Test direct GetKey
			key3, err := store.GetKey(account)
			if err != nil {
				t.Fatalf("GetKey() failed: %v", err)
			}

			for i, b := range key1 {
				if key3[i] != b {
					t.Errorf("GetKey() returned different key")
					break
				}
			}

			// Test SetKey with new key
			newKey := make([]byte, crypto.KeySize)
			rand.Read(newKey)

			err = store.SetKey(account+"2", newKey)
			if err != nil {
				t.Fatalf("SetKey() failed: %v", err)
			}

			// Verify the new key was stored
			retrievedKey, err := store.GetKey(account + "2")
			if err != nil {
				t.Fatalf("GetKey() after SetKey() failed: %v", err)
			}

			for i, b := range newKey {
				if retrievedKey[i] != b {
					t.Errorf("SetKey() didn't store correct key")
					break
				}
			}
		})
	}
}

func BenchmarkMockKeyStore_LoadOrCreateKey(b *testing.B) {
	store := newMockKeyStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		account := "bench-test"
		_, err := store.LoadOrCreateKey(account)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMockKeyStore_GetKey(b *testing.B) {
	store := newMockKeyStore()
	account := "bench-test"

	// Setup
	_, err := store.CreateKey(account)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.GetKey(account)
		if err != nil {
			b.Fatal(err)
		}
	}
}
