//go:build !darwin

package keystore

import (
	"errors"
)

// setGenericPassword is a fallback implementation for non-macOS systems
// In a real-world scenario, you might want to use a different secure storage mechanism
func setGenericPassword(label, service, account string, password []byte) error {
	// For non-macOS systems, we can't use the Keychain
	// This is a placeholder that returns an error
	return errors.New("keychain storage not available on this platform")
}

// getGenericPassword is a fallback implementation for non-macOS systems
func getGenericPassword(service, account string) (username string, password []byte, err error) {
	// For non-macOS systems, we can't use the Keychain
	// This is a placeholder that returns an error
	return "", nil, errors.New("keychain storage not available on this platform")
}
