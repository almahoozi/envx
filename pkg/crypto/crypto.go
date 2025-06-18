package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const (
	MagicPrefix = "envx"
	KeySize     = 32 // 256-bit key
)

// Encryptor defines the interface for encryption operations
type Encryptor interface {
	Encrypt(plaintext string, key []byte) (string, error)
	Decrypt(ciphertext string, key []byte) (string, error)
	IsEncrypted(value string) bool
}

// AESEncryptor implements the Encryptor interface using AES-GCM
type AESEncryptor struct{}

// NewAESEncryptor creates a new AES encryptor
func NewAESEncryptor() *AESEncryptor {
	return &AESEncryptor{}
}

// Encrypt encrypts a plaintext string using AES-GCM encryption
func (e *AESEncryptor) Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != KeySize {
		return "", fmt.Errorf("invalid key size: expected %d bytes, got %d", KeySize, len(key))
	}

	// If value is already encrypted, don't encrypt it again
	if e.IsEncrypted(plaintext) {
		return plaintext, nil
	}

	plaintextBytes := []byte(plaintext)
	ciphertext, err := e.encryptAES(key, plaintextBytes)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	// Prepend magic bytes to identify encrypted data
	ciphertext = append([]byte(MagicPrefix), ciphertext...)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a ciphertext string using AES-GCM decryption
func (e *AESEncryptor) Decrypt(ciphertext string, key []byte) (string, error) {
	if len(key) != KeySize {
		return "", fmt.Errorf("invalid key size: expected %d bytes, got %d", KeySize, len(key))
	}

	// First try to decode the value as base64, if that fails it's not encrypted
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return ciphertext, nil // Not base64, return as-is
	}

	// Check for magic prefix
	if len(decoded) < len(MagicPrefix) || !strings.HasPrefix(string(decoded), MagicPrefix) {
		return ciphertext, nil // No magic prefix, treat as unencrypted
	}

	// Decrypt the data after removing magic prefix
	plaintext, err := e.decryptAES(key, decoded[len(MagicPrefix):])
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a value appears to be encrypted
func (e *AESEncryptor) IsEncrypted(value string) bool {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return false
	}

	return len(decoded) > len(MagicPrefix) && strings.HasPrefix(string(decoded), MagicPrefix)
}

// encryptAES performs AES-GCM encryption
func (e *AESEncryptor) encryptAES(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAES performs AES-GCM decryption
func (e *AESEncryptor) decryptAES(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
} 