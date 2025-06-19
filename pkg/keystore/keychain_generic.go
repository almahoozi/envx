//go:build !darwin

package keystore

import (
	"errors"
	"os"
	"runtime"
)

// setGenericPassword is a fallback implementation for non-macOS systems
func setGenericPassword(label, service, account string, password []byte) error {
	// For non-macOS systems, we can't use the Keychain
	// This is a placeholder that returns an error
	if runtime.GOOS != "darwin" && os.Getenv("CI") == "" {
		return errors.New("keychain storage not available on this platform - use macOS or consider alternative secure storage")
	}
	return errors.New("keychain storage not available on this platform")
}

// getGenericPassword is a fallback implementation for non-macOS systems
func getGenericPassword(service, account string) (username string, password []byte, err error) {
	// For non-macOS systems, we can't use the Keychain
	// This is a placeholder that returns an error
	if runtime.GOOS != "darwin" && os.Getenv("CI") == "" {
		return "", nil, errors.New("keychain storage not available on this platform - use macOS or consider alternative secure storage")
	}
	return "", nil, errors.New("keychain storage not available on this platform")
}
