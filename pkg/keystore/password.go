package keystore

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"

	"crypto/pbkdf2"

	"github.com/almahoozi/envx/pkg/crypto"
	"golang.org/x/term"
)

const (
	// PBKDF2 parameters
	DefaultIterations = 100000 // OWASP recommended minimum
	SaltSize          = 32     // 256-bit salt
)

// PasswordKeyStore implements KeyStore using password-based key derivation
// Keys are derived from passwords on-demand and never stored
type PasswordKeyStore struct {
	iterations int
	promptFunc func(string) (string, error) // For dependency injection in tests
	password   string                       // Optional: password for non-interactive use
}

// PasswordKeyStoreConfig holds configuration for password-based keystore
type PasswordKeyStoreConfig struct {
	Iterations int
	PromptFunc func(string) (string, error)
	Password   string // Optional: password for non-interactive use
}

// NewPasswordKeyStore creates a new password-based keystore
func NewPasswordKeyStore(config *PasswordKeyStoreConfig) KeyStore {
	iterations := DefaultIterations
	promptFunc := promptForPassword
	password := ""

	if config != nil {
		if config.Iterations > 0 {
			iterations = config.Iterations
		}
		if config.PromptFunc != nil {
			promptFunc = config.PromptFunc
		}
		if config.Password != "" {
			password = config.Password
		}
	}

	return &PasswordKeyStore{
		iterations: iterations,
		promptFunc: promptFunc,
		password:   password,
	}
}

// GetKey derives a key from password (password is prompted from user or taken from config)
func (p *PasswordKeyStore) GetKey(account string) ([]byte, error) {
	password, err := p.getPassword(fmt.Sprintf("Enter password for %s", account))
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}

	salt, err := p.getSalt(account)
	if err != nil {
		return nil, fmt.Errorf("failed to get salt: %w", err)
	}

	key, err := pbkdf2.Key(sha256.New, password, salt, p.iterations, crypto.KeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	return key, nil
}

// SetKey is not applicable for password-based keystore - passwords are not stored
func (p *PasswordKeyStore) SetKey(account string, key []byte) error {
	return fmt.Errorf("SetKey not supported for password-based keystore - keys are derived from passwords")
}

// CreateKey creates a new salt for the account and prompts for password
func (p *PasswordKeyStore) CreateKey(account string) ([]byte, error) {
	// Generate a new salt
	salt := make([]byte, SaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Store the salt
	err = p.setSalt(account, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to store salt: %w", err)
	}

	// Get password (from env var, config, or prompt)
	password, err := p.getPassword(fmt.Sprintf("Create password for %s", account))
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}

	key, err := pbkdf2.Key(sha256.New, password, salt, p.iterations, crypto.KeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	return key, nil
}

// LoadOrCreateKey attempts to load salt and derive key, or creates new salt if it doesn't exist
func (p *PasswordKeyStore) LoadOrCreateKey(account string) ([]byte, error) {
	// Check if salt exists
	_, err := p.getSalt(account)
	if err == nil {
		// Salt exists, derive key from password
		return p.GetKey(account)
	}

	// Check if the error is specifically because the file doesn't exist
	saltFile := p.getSaltFilePath(account)
	if _, statErr := os.Stat(saltFile); os.IsNotExist(statErr) {
		// Salt file doesn't exist, create new one
		return p.CreateKey(account)
	}

	// Salt file exists but there was another error reading it - don't overwrite
	return nil, fmt.Errorf("salt file exists but cannot be read: %w", err)
}

// getSalt retrieves the salt for an account from a file
func (p *PasswordKeyStore) getSalt(account string) ([]byte, error) {
	saltFile := p.getSaltFilePath(account)
	salt, err := os.ReadFile(saltFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read salt file: %w", err)
	}

	if len(salt) != SaltSize {
		return nil, fmt.Errorf("invalid salt size: expected %d bytes, got %d", SaltSize, len(salt))
	}

	return salt, nil
}

// setSalt stores the salt for an account to a file
func (p *PasswordKeyStore) setSalt(account string, salt []byte) error {
	if len(salt) != SaltSize {
		return fmt.Errorf("invalid salt size: expected %d bytes, got %d", SaltSize, len(salt))
	}

	saltFile := p.getSaltFilePath(account)

	// Create directory if it doesn't exist
	saltDir := getSaltDir()
	err := os.MkdirAll(saltDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create salt directory: %w", err)
	}

	err = os.WriteFile(saltFile, salt, 0600)
	if err != nil {
		return fmt.Errorf("failed to write salt file: %w", err)
	}

	return nil
}

// getSaltFilePath returns the file path for storing the salt
func (p *PasswordKeyStore) getSaltFilePath(account string) string {
	return fmt.Sprintf("%s/%s.salt", getSaltDir(), account)
}

// getSaltDir is a variable function that returns the directory for storing salt files
// This allows for easy testing by reassigning the function
var getSaltDir = func() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory is not available
		return ".envx/salts"
	}
	return fmt.Sprintf("%s/.config/envx/salts", homeDir)
}

// promptForPassword prompts the user for a password without echoing to terminal
func promptForPassword(prompt string) (string, error) {
	fmt.Printf("%s: ", prompt)

	// Read password without echoing to terminal
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Println() // Print newline after password input
	return string(bytePassword), nil
}

// getPassword returns the configured password or prompts for one
func (p *PasswordKeyStore) getPassword(prompt string) (string, error) {
	// Check environment variable first
	if envPassword := os.Getenv("ENVX_PASSWORD"); envPassword != "" {
		return envPassword, nil
	}

	// Use configured password if available
	if p.password != "" {
		return p.password, nil
	}

	// Fall back to prompting
	return p.promptFunc(prompt)
}
