package keystore

import (
	"crypto/rand"
	"fmt"

	"github.com/almahoozi/envx/pkg/crypto"
)

// KeyStore defines the interface for key storage operations
type KeyStore interface {
	GetKey(account string) ([]byte, error)
	SetKey(account string, key []byte) error
	CreateKey(account string) ([]byte, error)
	LoadOrCreateKey(account string) ([]byte, error)
}

// Config holds keystore configuration
type Config struct {
	App     string
	Service string
}

// DefaultConfig returns the default keystore configuration
func DefaultConfig() *Config {
	return &Config{
		App:     "envx",
		Service: "com.almahoozi.envx",
	}
}

// macOSKeyStore implements KeyStore using macOS Keychain
type macOSKeyStore struct {
	config *Config
}

// NewMacOSKeyStore creates a new macOS keystore instance
func NewMacOSKeyStore(config *Config) KeyStore {
	if config == nil {
		config = DefaultConfig()
	}
	return &macOSKeyStore{config: config}
}

// GetKey retrieves a key from the macOS Keychain
func (k *macOSKeyStore) GetKey(account string) ([]byte, error) {
	_, key, err := getGenericPassword(k.config.Service, account)
	if err != nil {
		return nil, fmt.Errorf("failed to get key from keychain: %w", err)
	}
	return key, nil
}

// SetKey stores a key in the macOS Keychain
func (k *macOSKeyStore) SetKey(account string, key []byte) error {
	if len(key) != crypto.KeySize {
		return fmt.Errorf("invalid key size: expected %d bytes, got %d", crypto.KeySize, len(key))
	}

	err := setGenericPassword(k.config.App, k.config.Service, account, key)
	if err != nil {
		return fmt.Errorf("failed to set key in keychain: %w", err)
	}
	return nil
}

// CreateKey generates a new random key and stores it in the keychain
func (k *macOSKeyStore) CreateKey(account string) ([]byte, error) {
	key := make([]byte, crypto.KeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	err = k.SetKey(account, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// LoadOrCreateKey attempts to load an existing key, or creates a new one if it doesn't exist
func (k *macOSKeyStore) LoadOrCreateKey(account string) ([]byte, error) {
	// Try to get existing key
	key, err := k.GetKey(account)
	if err == nil && len(key) == crypto.KeySize {
		return key, nil
	}

	// Key doesn't exist or is invalid, create a new one
	key, err = k.CreateKey(account)
	if err != nil {
		return nil, fmt.Errorf("failed to create new key: %w", err)
	}

	return key, nil
}
