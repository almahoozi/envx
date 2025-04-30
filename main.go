package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
)

//go:embed envx.1
var man string

type envVar struct {
	key   string
	value string
}

func main() {
	if err := start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func loadEnv(name string, decryptionKey []byte) (vars []envVar, err error) {
	file, err := os.Open(name)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error opening .env file:", err)
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// TODO: Support setting empty values
			continue
		}
		// TODO: Support decoding base64 values
		// TODO: Support decrypting encrypted values. Ask for password if any encrypted value is present.
		// Can we make this a key-chain like experience? Or like just ask for the terminal user's password?
		key, value := parts[0], parts[1]
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// If value is enclosed in double quotes, remove them
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		}

		vars = append(vars, envVar{key, value})
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading .env file:", err)
	}

	return vars, nil
}

func loadDecryptedEnv(name string, decryptionKey []byte) (vars []envVar, err error) {
	file, err := os.Open(name)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error opening .env file:", err)
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// TODO: Support setting empty values
			continue
		}
		// TODO: Support decoding base64 values
		// TODO: Support decrypting encrypted values. Ask for password if any encrypted value is present.
		// Can we make this a key-chain like experience? Or like just ask for the terminal user's password?
		key, value := parts[0], parts[1]
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// If value is enclosed in double quotes, remove them
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		}

		plaintext, err := decrypt(value, decryptionKey)
		if err != nil {
			fmt.Println("Error decrypting value:", err)
			return nil, err
		}

		vars = append(vars, envVar{key, plaintext})
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading .env file:", err)
	}

	return vars, nil
}

func decryptAES(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, ciphciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func decrypt(value string, key []byte) (string, error) {
	// First try to decode the value as base64, if that's not the case then it's not encrypted
	ciphertext, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return value, nil
	}

	// if we encrypted it it would start with the bytes for "envx"
	if !strings.HasPrefix(string(ciphertext), "envx") {
		// It could be just base 64 encoded in order to have multiple lines or special chars,
		// but there's no way to know for sure what the intention is
		return value, nil
	}

	// Key is a 256-bit aes key
	// The first 4 bytes are the magic bytes "envx"
	// The next 16 bytes are the IV
	// The rest is the encrypted data

	plaintext, err := decryptAES(key, ciphertext[4:])
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func encryptAES(key, plaintext []byte) ([]byte, error) {
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
		log.Fatal(err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func encrypt(value string, key []byte) (string, error) {
	// If value is already encrypted, don't encrypt it again
	if strings.HasPrefix(value, "envx") {
		return value, nil
	}

	plaintext := []byte(value)
	ciphertext, err := encryptAES(key, plaintext)
	if err != nil {
		fmt.Println("Error encrypting value:", err)
		return value, err
	}

	// TODO: Instead of prepending envx directly, prepend a random byte which is
	// used to "shift" the next 4 bytes (envx) by that amount. This way we can
	// have a different "magic" bytes for each encryption.
	// We could extend this to be 4 random bytes but this doesn't really add any
	// cryptographic value.

	ciphertext = append([]byte("envx"), ciphertext...)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

const (
	app     = "envx"
	service = "com.almahoozi.envx"
)

func loadKey() ([]byte, error) {
	user, err := user.Current()
	if err != nil {
		log.Fatal("Error getting current user:", err)
	}

	var key []byte
	if key, err = getKey(user.Username); err != nil {
		// if it simply doesn't exist create it
		// FIX: We may error due to other reasons! We shouldn't overwrite the key!
		// For example if we don't have permissions to access the keychain or the key.
		log.Println("Key not found, creating new key")
	}

	if len(key) == 0 {
		err = createKey(user.Username)
		if err != nil {
			log.Fatal("Error storing key:", err)
		}

		fmt.Println("Key stored in macOS Keychain")

		key, err = getKey(user.Username)
		if err != nil {
			log.Fatal("Error fetching key:", err)
		}
	}
	return key, nil
}

func getKey(account string) ([]byte, error) {
	_, key, err := getGenericPassword(service, account)
	return key, err
}

func createKey(account string) error {
	key := make([]byte, 32) // 256-bit key
	_, err := rand.Read(key)
	if err != nil {
		return err
	}

	return setGenericPassword(app, service, account, key)
}
