package env

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/almahoozi/envx/pkg/crypto"
	"github.com/almahoozi/envx/pkg/errlog"
)

// Variable represents an environment variable key-value pair
type Variable struct {
	Key   string
	Value string
}

// Variables is a slice of Variable
type Variables []Variable

// ToMap converts Variables to a map[string]string
func (vars Variables) ToMap() map[string]string {
	result := make(map[string]string, len(vars))
	for _, v := range vars {
		result[v.Key] = v.Value
	}
	return result
}

// Get returns the variable with the given key, or nil if not found
func (vars Variables) Get(key string) *Variable {
	for i, v := range vars {
		if v.Key == key {
			return &vars[i]
		}
	}
	return nil
}

// Set updates an existing variable or adds a new one
func (vars *Variables) Set(key, value string) {
	for i, v := range *vars {
		if v.Key == key {
			(*vars)[i].Value = value
			return
		}
	}
	*vars = append(*vars, Variable{Key: key, Value: value})
}

// Remove removes a variable by key
func (vars *Variables) Remove(key string) bool {
	for i, v := range *vars {
		if v.Key == key {
			*vars = append((*vars)[:i], (*vars)[i+1:]...)
			return true
		}
	}
	return false
}

// Loader defines the interface for loading environment variables
type Loader interface {
	Load(ctx context.Context, filename string) (Variables, error)
	LoadWithDecryption(ctx context.Context, filename string, encryptor crypto.Encryptor, key []byte) (Variables, error)
}

// Writer defines the interface for writing environment variables
type Writer interface {
	Write(filename string, vars Variables, format Format) error
}

// Format represents the output format for environment variables
type Format string

const (
	FormatEnv  Format = "env"
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// FileLoader implements Loader for loading from files
type FileLoader struct{}

// NewFileLoader creates a new file loader
func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

// Load loads environment variables from a file
func (l *FileLoader) Load(ctx context.Context, filename string) (Variables, error) {
	file, err := os.Open(filename) // #nosec G304 -- User-provided filename is intentional for env file loading
	if err != nil {
		if os.IsNotExist(err) {
			return Variables{}, nil // Return empty variables if file doesn't exist
		}
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer errlog.FnLog(ctx, file.Close)

	// Pre-allocate with reasonable capacity to reduce reallocations
	vars := make(Variables, 0, 32)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments without trimming first
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Find the first '=' character
		eqIndex := strings.IndexByte(line, '=')
		if eqIndex == -1 {
			continue // Skip malformed lines
		}

		// Extract key and value with minimal allocations
		key := strings.TrimSpace(line[:eqIndex])
		if len(key) == 0 {
			continue // Skip lines with empty keys
		}

		value := strings.TrimSpace(line[eqIndex+1:])

		// Remove quotes if present (optimized check)
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		vars = append(vars, Variable{Key: key, Value: value})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}

	return vars, nil
}

// LoadWithDecryption loads and decrypts environment variables from a file
func (l *FileLoader) LoadWithDecryption(ctx context.Context, filename string, encryptor crypto.Encryptor, key []byte) (Variables, error) {
	vars, err := l.Load(ctx, filename)
	if err != nil {
		return nil, err
	}

	// Decrypt all values
	for i, v := range vars {
		decrypted, err := encryptor.Decrypt(v.Value, key)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt variable %s: %w", v.Key, err)
		}
		vars[i].Value = decrypted
	}

	return vars, nil
}

// FileWriter implements Writer for writing to files
type FileWriter struct{}

// NewFileWriter creates a new file writer
func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

// Write writes environment variables to a file in the specified format
func (w *FileWriter) Write(filename string, vars Variables, format Format) error {
	var content string
	switch format {
	case FormatJSON:
		content = w.formatJSON(vars)
	case FormatYAML:
		return fmt.Errorf("YAML format not yet implemented")
	default:
		content = w.formatEnv(vars)
	}

	err := os.WriteFile(filename, []byte(content), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}

// formatEnv formats variables as .env format
func (w *FileWriter) formatEnv(vars Variables) string {
	var sb strings.Builder
	for _, v := range vars {
		// Quote values that contain spaces or special characters
		value := v.Value
		if strings.ContainsAny(value, " \t\n\"") {
			value = fmt.Sprintf("%q", value)
		}
		sb.WriteString(fmt.Sprintf("%s=%s\n", v.Key, value))
	}
	return sb.String()
}

// formatJSON formats variables as JSON
func (w *FileWriter) formatJSON(vars Variables) string {
	var parts []string
	for _, v := range vars {
		parts = append(parts, fmt.Sprintf("%q:%q", v.Key, v.Value))
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ","))
}

// BuildFilename constructs a filename based on base file and optional name suffix
func BuildFilename(baseFile, name string) string {
	if name == "" {
		return baseFile
	}
	return fmt.Sprintf("%s.%s", baseFile, name)
}
