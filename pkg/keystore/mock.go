package keystore

import (
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/almahoozi/envx/pkg/crypto"
)

// MockKeyStore is an in-memory keystore for testing
type MockKeyStore struct {
	keys map[string][]byte
	mu   sync.RWMutex
}

// NewMockKeyStore creates a new mock keystore
func NewMockKeyStore() KeyStore {
	return &MockKeyStore{
		keys: make(map[string][]byte),
	}
}

// GetKey retrieves a key from the mock store
func (m *MockKeyStore) GetKey(account string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key, exists := m.keys[account]
	if !exists {
		return nil, fmt.Errorf("key not found for account: %s", account)
	}

	return key, nil
}

// SetKey stores a key in the mock store
func (m *MockKeyStore) SetKey(account string, key []byte) error {
	if len(key) != crypto.KeySize {
		return fmt.Errorf("invalid key size: expected %d bytes, got %d", crypto.KeySize, len(key))
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.keys[account] = make([]byte, len(key))
	copy(m.keys[account], key)

	return nil
}

// CreateKey generates a new random key and stores it
func (m *MockKeyStore) CreateKey(account string) ([]byte, error) {
	key := make([]byte, crypto.KeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	err = m.SetKey(account, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// LoadOrCreateKey attempts to load an existing key, or creates a new one if it doesn't exist
func (m *MockKeyStore) LoadOrCreateKey(account string) ([]byte, error) {
	// Try to get existing key
	key, err := m.GetKey(account)
	if err == nil && len(key) == crypto.KeySize {
		return key, nil
	}

	// Key doesn't exist or is invalid, create a new one
	key, err = m.CreateKey(account)
	if err != nil {
		return nil, fmt.Errorf("failed to create new key: %w", err)
	}

	return key, nil
}
