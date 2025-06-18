package crypto

import (
	"crypto/rand"
	"strings"
	"testing"
)

func TestNewAESEncryptor(t *testing.T) {
	encryptor := NewAESEncryptor()
	if encryptor == nil {
		t.Fatal("NewAESEncryptor returned nil")
	}
}

func TestAESEncryptor_Encrypt(t *testing.T) {
	encryptor := NewAESEncryptor()
	key := make([]byte, KeySize)
	rand.Read(key)

	tests := []struct {
		name      string
		plaintext string
		key       []byte
		wantErr   bool
	}{
		{
			name:      "valid encryption",
			plaintext: "hello world",
			key:       key,
			wantErr:   false,
		},
		{
			name:      "empty plaintext",
			plaintext: "",
			key:       key,
			wantErr:   false,
		},
		{
			name:      "long plaintext",
			plaintext: strings.Repeat("a", 1000),
			key:       key,
			wantErr:   false,
		},
		{
			name:      "special characters",
			plaintext: "Hello 世界! @#$%^&*()",
			key:       key,
			wantErr:   false,
		},
		{
			name:      "invalid key size - too short",
			plaintext: "test",
			key:       make([]byte, 16), // 128-bit key instead of 256-bit
			wantErr:   true,
		},
		{
			name:      "invalid key size - too long",
			plaintext: "test",
			key:       make([]byte, 64), // 512-bit key instead of 256-bit
			wantErr:   true,
		},
		{
			name:      "nil key",
			plaintext: "test",
			key:       nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encryptor.Encrypt(tt.plaintext, tt.key)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Encrypt() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Encrypt() unexpected error: %v", err)
				return
			}

			// Check that result is base64 encoded
			if result == "" {
				t.Errorf("Encrypt() returned empty string")
			}

			// Check that result is different from plaintext (unless already encrypted)
			if result == tt.plaintext && !encryptor.IsEncrypted(tt.plaintext) {
				t.Errorf("Encrypt() returned same as plaintext")
			}

			// Check that result can be identified as encrypted
			if !encryptor.IsEncrypted(result) {
				t.Errorf("Encrypt() result is not identified as encrypted")
			}
		})
	}
}

func TestAESEncryptor_Decrypt(t *testing.T) {
	encryptor := NewAESEncryptor()
	key := make([]byte, KeySize)
	rand.Read(key)

	tests := []struct {
		name      string
		plaintext string
		key       []byte
		wantErr   bool
		setupFunc func(string, []byte) string // function to setup ciphertext
	}{
		{
			name:      "valid decryption",
			plaintext: "hello world",
			key:       key,
			wantErr:   false,
			setupFunc: func(plain string, k []byte) string {
				encrypted, _ := encryptor.Encrypt(plain, k)
				return encrypted
			},
		},
		{
			name:      "empty plaintext",
			plaintext: "",
			key:       key,
			wantErr:   false,
			setupFunc: func(plain string, k []byte) string {
				encrypted, _ := encryptor.Encrypt(plain, k)
				return encrypted
			},
		},
		{
			name:      "long plaintext",
			plaintext: strings.Repeat("test data ", 100),
			key:       key,
			wantErr:   false,
			setupFunc: func(plain string, k []byte) string {
				encrypted, _ := encryptor.Encrypt(plain, k)
				return encrypted
			},
		},
		{
			name:      "unencrypted data returns as-is",
			plaintext: "plain text",
			key:       key,
			wantErr:   false,
			setupFunc: func(plain string, k []byte) string {
				return plain // return plaintext as-is
			},
		},
		{
			name:      "invalid base64 returns as-is",
			plaintext: "invalid base64!!!",
			key:       key,
			wantErr:   false,
			setupFunc: func(plain string, k []byte) string {
				return plain
			},
		},
		{
			name:      "invalid key size",
			plaintext: "test",
			key:       make([]byte, 16),
			wantErr:   true,
			setupFunc: func(plain string, k []byte) string {
				// Use valid key for encryption, invalid for decryption
				validKey := make([]byte, KeySize)
				rand.Read(validKey)
				encrypted, _ := encryptor.Encrypt(plain, validKey)
				return encrypted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext := tt.setupFunc(tt.plaintext, tt.key)

			result, err := encryptor.Decrypt(ciphertext, tt.key)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Decrypt() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Decrypt() unexpected error: %v", err)
				return
			}

			if result != tt.plaintext {
				t.Errorf("Decrypt() = %q, want %q", result, tt.plaintext)
			}
		})
	}
}

func TestAESEncryptor_IsEncrypted(t *testing.T) {
	encryptor := NewAESEncryptor()
	key := make([]byte, KeySize)
	rand.Read(key)

	// Test unencrypted data
	unencryptedData := []string{
		"plain text",
		"",
		"not base64 encoded",
		"dmFsaWQgYmFzZTY0IGJ1dCBub3QgZW52eA==", // valid base64 but doesn't start with "envx"
	}

	for _, data := range unencryptedData {
		t.Run("unencrypted_"+data, func(t *testing.T) {
			if encryptor.IsEncrypted(data) {
				t.Errorf("IsEncrypted(%q) = true, want false", data)
			}
		})
	}

	// Test encrypted data
	plaintexts := []string{
		"hello world",
		"",
		"special chars: 世界!@#$%",
		strings.Repeat("long text ", 50),
	}

	for _, plaintext := range plaintexts {
		t.Run("encrypted_"+plaintext, func(t *testing.T) {
			encrypted, err := encryptor.Encrypt(plaintext, key)
			if err != nil {
				t.Fatalf("Failed to encrypt test data: %v", err)
			}

			if !encryptor.IsEncrypted(encrypted) {
				t.Errorf("IsEncrypted(%q) = false, want true", encrypted)
			}
		})
	}
}

func TestAESEncryptor_EncryptDecryptRoundTrip(t *testing.T) {
	encryptor := NewAESEncryptor()
	key := make([]byte, KeySize)
	rand.Read(key)

	testCases := []string{
		"hello world",
		"",
		"special characters: 世界!@#$%^&*()",
		strings.Repeat("long text ", 100),
		"multiline\ntext\nwith\ttabs",
		`JSON: {"key": "value", "number": 123}`,
		"SQL: SELECT * FROM users WHERE id = 1;",
	}

	for _, plaintext := range testCases {
		t.Run("roundtrip_"+plaintext, func(t *testing.T) {
			// Encrypt
			encrypted, err := encryptor.Encrypt(plaintext, key)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Verify it's marked as encrypted
			if !encryptor.IsEncrypted(encrypted) {
				t.Errorf("Encrypted data not identified as encrypted")
			}

			// Decrypt
			decrypted, err := encryptor.Decrypt(encrypted, key)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// Verify round-trip
			if decrypted != plaintext {
				t.Errorf("Round-trip failed: got %q, want %q", decrypted, plaintext)
			}
		})
	}
}

func TestAESEncryptor_EncryptTwiceSameResult(t *testing.T) {
	encryptor := NewAESEncryptor()
	key := make([]byte, KeySize)
	rand.Read(key)

	plaintext := "test data"

	// First encryption
	encrypted1, err := encryptor.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	// Second encryption of already encrypted data should return same result
	encrypted2, err := encryptor.Encrypt(encrypted1, key)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	if encrypted1 != encrypted2 {
		t.Errorf("Encrypting already encrypted data changed result: %q != %q", encrypted1, encrypted2)
	}
}

func TestAESEncryptor_DifferentKeysProduceDifferentResults(t *testing.T) {
	encryptor := NewAESEncryptor()

	key1 := make([]byte, KeySize)
	key2 := make([]byte, KeySize)
	rand.Read(key1)
	rand.Read(key2)

	plaintext := "test data"

	encrypted1, err := encryptor.Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("Encryption with key1 failed: %v", err)
	}

	encrypted2, err := encryptor.Encrypt(plaintext, key2)
	if err != nil {
		t.Fatalf("Encryption with key2 failed: %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Errorf("Different keys produced same encrypted result")
	}

	// Verify cross-decryption fails
	_, err = encryptor.Decrypt(encrypted1, key2)
	if err == nil {
		t.Errorf("Decryption with wrong key should have failed")
	}

	_, err = encryptor.Decrypt(encrypted2, key1)
	if err == nil {
		t.Errorf("Decryption with wrong key should have failed")
	}
}

func TestConstants(t *testing.T) {
	if MagicPrefix != "envx" {
		t.Errorf("MagicPrefix = %q, want %q", MagicPrefix, "envx")
	}

	if KeySize != 32 {
		t.Errorf("KeySize = %d, want %d", KeySize, 32)
	}
}

// Benchmark tests
func BenchmarkAESEncryptor_Encrypt(b *testing.B) {
	encryptor := NewAESEncryptor()
	key := make([]byte, KeySize)
	rand.Read(key)
	plaintext := "benchmark test data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Encrypt(plaintext, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAESEncryptor_Decrypt(b *testing.B) {
	encryptor := NewAESEncryptor()
	key := make([]byte, KeySize)
	rand.Read(key)
	plaintext := "benchmark test data"

	encrypted, err := encryptor.Encrypt(plaintext, key)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Decrypt(encrypted, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}
