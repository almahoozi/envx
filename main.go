package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"syscall"
)

//go:embed envx.1
var man string

type envVar struct {
	key   string
	value string
}

func arg(i int) string {
	if len(os.Args) > i {
		return os.Args[i]
	}
	return ""
}

func main() {
	if len(os.Args) < 2 {
		name := arg(0)
		fmt.Printf("Usage: %s <executable> [args]\n", name)
		// TODO: Support options (ex: selecting the env file --env=dev|-e dev selects .env.dev)
		// --file=x.env|-f x.env selects x.env
		// --password|-p allows to set a password for encrypted values
		// fmt.Printf("Usage: %s [options] <executable> [args]\n", name)
		os.Exit(1)
	}

	i := 1

	if cmd := arg(i); cmd == "help" || cmd == "man" {
		log.Println(man)
	}

	// optimistically load encryption key
	key, err := loadKey()
	if err != nil {
		log.Fatal("Error loading key:", err)
	}

	// Load .env file if exists
	vars := loadEnv(key)

	// If first arg is "encrypt" then encrypt the value
	if arg(i) == "encrypt" {
		var write bool
		i++
		if arg(i) == "-w" {
			write = true
			i++
		}

		for i, v := range vars {
			// If it isn't already encrypted, encrypt it
			// Well, we decrypted everything, so just encrypt everything
			vars[i].value = encrypt(v.value, key)
		}

		output := ""
		for _, v := range vars {
			output += fmt.Sprintf("%s=%s\n", v.key, v.value)
		}

		if write {
			err := os.WriteFile(".env", []byte(output), 0o644)
			if err != nil {
				fmt.Println("Error writing .env file:", err)
			}
		} else {
			fmt.Println(output)
		}

		os.Exit(0)
	}

	execPath := arg(i)
	args := os.Args[i:] // Pass all args as they are

	for _, v := range vars {
		os.Setenv(v.key, v.value)
	}

	// Execute the new process in place of the Go process
	// TODO: execpath could be in the $PATH, not in the current directory
	execPath, err = exec.LookPath(execPath)
	if err != nil {
		fmt.Println("Executable not found:", execPath)
		os.Exit(1)
	}

	err = syscall.Exec(execPath, args, os.Environ())
	if err != nil {
		fmt.Println("Error executing process:", err, execPath, args)
		os.Exit(1)
	}
}

func loadEnv(decryptionKey []byte) (vars []envVar) {
	name := ".env"

	file, err := os.Open(name)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error opening .env file:", err)
		}
		return nil
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
		if plaintext, err := decrypt(value, decryptionKey); err != nil {
			fmt.Println("Error decrypting value:", err)
		} else {
			value = plaintext
		}

		vars = append(vars, envVar{key, value})
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading .env file:", err)
	}

	return vars
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

func encrypt(value string, key []byte) string {
	// If value is already encrypted, don't encrypt it again
	if strings.HasPrefix(value, "envx") {
		return value
	}

	plaintext := []byte(value)
	ciphertext, err := encryptAES(key, plaintext)
	if err != nil {
		fmt.Println("Error encrypting value:", err)
		return value
	}

	// TODO: Instead of prepending envx directly, prepend a random byte which is
	// used to "shift" the next 4 bytes (envx) by that amount. This way we can
	// have a different "magic" bytes for each encryption.
	// We could extend this to be 4 random bytes but this doesn't really add any
	// cryptographic value.

	ciphertext = append([]byte("envx"), ciphertext...)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

const (
	app     = "envx"
	service = "com.almahoozi.envx"
	account = "default"
)

func loadKey() ([]byte, error) {
	user, err := user.Current()
	if err != nil {
		log.Fatal("Error getting current user:", err)
	}

	var key []byte
	if key, err = getKey(user.Username); err != nil {
		// if it simply doesn't exist create it
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
	cmd := exec.Command("security", "find-generic-password",
		"-s", service,
		"-a", account,
		"-w")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(strings.TrimSpace(string(out)))
	return key, err
}

func createKey(account string) error {
	key := make([]byte, 32) // 256-bit key
	_, err := rand.Read(key)
	if err != nil {
		return err
	}

	cmd := exec.Command("security", "add-generic-password",
		"-l", app,
		"-s", service,
		"-a", account,
		"-X", hex.EncodeToString(key),
		//"-T", "/Users/hussam/Documents/Source/personal/keychain/envxx",
	)
	return cmd.Run()
}
