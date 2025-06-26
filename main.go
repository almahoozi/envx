package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/user"

	"github.com/almahoozi/envx/pkg/crypto"
	"github.com/almahoozi/envx/pkg/env"
	"github.com/almahoozi/envx/pkg/keystore"
)

// testKeystoreConfig can be set during tests to use a different keychain item
var testKeystoreConfig *keystore.Config

// testKeystore can be set during tests to use a mock keystore
var testKeystore keystore.KeyStore

//go:embed envx.1
var man string

func main() {
	if err := start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// loadEnv loads environment variables from a file using the env package
func loadEnv(ctx context.Context, filename string) (env.Variables, error) {
	loader := env.NewFileLoader()
	return loader.Load(ctx, filename)
}

// loadDecryptedEnv loads and decrypts environment variables from a file
func loadDecryptedEnv(ctx context.Context, filename string, encryptor crypto.Encryptor, key []byte) (env.Variables, error) {
	loader := env.NewFileLoader()
	return loader.LoadWithDecryption(ctx, filename, encryptor, key)
}

// KeyStoreType represents the type of keystore to use
type KeyStoreType string

const (
	KeyStoreTypeMacOS    KeyStoreType = "macos"
	KeyStoreTypePassword KeyStoreType = "password"
	KeyStoreTypeMock     KeyStoreType = "mock"
)

// avoid unused lint errors
var (
	_ = loadKey
	_ = loadKeyWithStringType
)

// loadKey loads or creates an encryption key using the default keystore
func loadKey() ([]byte, error) {
	return loadKeyWithType(KeyStoreTypeMacOS)
}

// loadKeyWithStringType loads or creates an encryption key using the specified keystore type string
func loadKeyWithStringType(storeTypeStr string) ([]byte, error) {
	return loadKeyWithStringTypeAndPassword(storeTypeStr, "")
}

// loadKeyWithStringTypeAndPassword loads or creates an encryption key using the specified keystore type string and password
func loadKeyWithStringTypeAndPassword(storeTypeStr, password string) ([]byte, error) {
	// As a workaround we set the password to byte(1) if it is empty using -P
	if password == emptyPassword {
		storeTypeStr = "password"
		password = ""
	}

	// Auto-detect password keystore if password is provided or ENVX_PASSWORD is set
	if password != "" || os.Getenv("ENVX_PASSWORD") != "" {
		storeTypeStr = "password"
	}

	storeType, err := parseKeyStoreType(storeTypeStr)
	if err != nil {
		return nil, err
	}
	return loadKeyWithTypeAndPassword(storeType, password)
}

// parseKeyStoreType converts a string to KeyStoreType
func parseKeyStoreType(storeTypeStr string) (KeyStoreType, error) {
	switch storeTypeStr {
	case "macos":
		return KeyStoreTypeMacOS, nil
	case "password":
		return KeyStoreTypePassword, nil
	case "mock":
		return KeyStoreTypeMock, nil
	default:
		return "", fmt.Errorf("unsupported keystore type: %s (supported: macos, password, mock)", storeTypeStr)
	}
}

// loadKeyWithType loads or creates an encryption key using the specified keystore type
func loadKeyWithType(storeType KeyStoreType) ([]byte, error) {
	return loadKeyWithTypeAndPassword(storeType, "")
}

// loadKeyWithTypeAndPassword loads or creates an encryption key using the specified keystore type and optional password
func loadKeyWithTypeAndPassword(storeType KeyStoreType, password string) ([]byte, error) {
	user, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	var store keystore.KeyStore
	if testKeystore != nil {
		// Use test keystore if set (for testing)
		store = testKeystore
	} else {
		// Use production keystore based on type
		switch storeType {
		case KeyStoreTypePassword:
			config := &keystore.PasswordKeyStoreConfig{
				Password: password,
			}
			store = keystore.NewPasswordKeyStore(config)
		case KeyStoreTypeMock:
			store = keystore.NewMockKeyStore()
		case KeyStoreTypeMacOS:
			fallthrough
		default:
			// Use test config if set (for testing), otherwise use default
			config := testKeystoreConfig
			store = keystore.NewMacOSKeyStore(config)
		}
	}

	key, err := store.LoadOrCreateKey(user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to load or create key: %w", err)
	}

	return key, nil
}
