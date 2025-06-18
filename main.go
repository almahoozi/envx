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

// loadKey loads or creates an encryption key using the default keystore
func loadKey() ([]byte, error) {
	// Use test config if set (for testing), otherwise use default
	config := testKeystoreConfig
	return loadKeyWithConfig(config)
}

// loadKeyWithConfig loads or creates an encryption key using the specified keystore config
func loadKeyWithConfig(config *keystore.Config) ([]byte, error) {
	user, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	store := keystore.NewMacOSKeyStore(config)
	key, err := store.LoadOrCreateKey(user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to load or create key: %w", err)
	}

	return key, nil
}
