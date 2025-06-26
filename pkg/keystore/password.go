package keystore

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/almahoozi/envx/pkg/crypto"
	"golang.org/x/crypto/pbkdf2"
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
}

// PasswordKeyStoreConfig holds configuration for password-based keystore
type PasswordKeyStoreConfig struct {
	Iterations int
	PromptFunc func(string) (string, error)
}

// NewPasswordKeyStore creates a new password-based keystore
func NewPasswordKeyStore(config *PasswordKeyStoreConfig) KeyStore {
	iterations := DefaultIterations
	promptFunc := promptForPassword

	if config != nil {
		if config.Iterations > 0 {
			iterations = config.Iterations
		}
		if config.PromptFunc != nil {
			promptFunc = config.PromptFunc
		}
	}

	return &PasswordKeyStore{
		iterations: iterations,
		promptFunc: promptFunc,
	}
}

// GetKey derives a key from password (password is prompted from user)
func (p *PasswordKeyStore) GetKey(account string) ([]byte, error) {
	password, err := p.promptFunc(fmt.Sprintf("Enter password for %s", account))
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}

	salt, err := p.getSalt(account)
	if err != nil {
		return nil, fmt.Errorf("failed to get salt: %w", err)
	}

	key := pbkdf2.Key([]byte(password), salt, p.iterations, crypto.KeySize, sha256.New)
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

	// Prompt for password and derive key
	password, err := p.promptFunc(fmt.Sprintf("Create password for %s", account))
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}

	// Confirm password
	confirmPassword, err := p.promptFunc(fmt.Sprintf("Confirm password for %s", account))
	if err != nil {
		return nil, fmt.Errorf("failed to confirm password: %w", err)
	}

	if password != confirmPassword {
		return nil, fmt.Errorf("passwords do not match")
	}

	key := pbkdf2.Key([]byte(password), salt, p.iterations, crypto.KeySize, sha256.New)
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

	// Salt doesn't exist, create new one
	return p.CreateKey(account)
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
