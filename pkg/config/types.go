package config

// Config represents the envx configuration
type Config struct {
	Keystore       string   `yaml:"keystore,omitempty"`
	File           string   `yaml:"file,omitempty"`
	Name           string   `yaml:"name,omitempty"`
	Format         string   `yaml:"format,omitempty"`
	KeyName        string   `yaml:"key_name,omitempty"`
	FileResolution []string `yaml:"file_resolution,omitempty"`
	BackupOnWrite  bool     `yaml:"backup_on_write,omitempty"`
}

// DefaultConfig returns the default configuration values
func DefaultConfig() *Config {
	return &Config{
		Keystore:       "macos",
		File:           ".env",
		Name:           "",
		Format:         "env",
		KeyName:        "envx.test",
		FileResolution: []string{".env"},
		BackupOnWrite:  true, // Default to true for safety
	}
}

// Merge merges the other config into this config, with other taking precedence
// Only non-empty values from other are merged
func (c *Config) Merge(other *Config) {
	if other.Keystore != "" {
		c.Keystore = other.Keystore
	}
	if other.File != "" {
		c.File = other.File
	}
	if other.Name != "" {
		c.Name = other.Name
	}
	if other.Format != "" {
		c.Format = other.Format
	}
	if other.KeyName != "" {
		c.KeyName = other.KeyName
	}
	if len(other.FileResolution) > 0 {
		c.FileResolution = other.FileResolution
	}
	// For boolean fields, we need to check if it was explicitly set
	// For now, we'll always merge the boolean value
	c.BackupOnWrite = other.BackupOnWrite
}

// ConfigSource represents where a configuration value came from
type ConfigSource string

const (
	SourceDefault   ConfigSource = "default"
	SourceGlobal    ConfigSource = "global"
	SourceDirectory ConfigSource = "directory"
	SourceCLI       ConfigSource = "cli"
	SourceEnv       ConfigSource = "env"
)

// ConfigValue represents a configuration value with its source
type ConfigValue struct {
	Value  string       `json:"value"`
	Source ConfigSource `json:"source"`
}

// ConfigArrayValue represents a configuration array value with its source
type ConfigArrayValue struct {
	Value  []string     `json:"value"`
	Source ConfigSource `json:"source"`
}

// ConfigReport represents the effective configuration with sources
type ConfigReport struct {
	Keystore       ConfigValue      `json:"keystore"`
	File           ConfigValue      `json:"file"`
	Name           ConfigValue      `json:"name"`
	Format         ConfigValue      `json:"format"`
	KeyName        ConfigValue      `json:"key_name"`
	FileResolution ConfigArrayValue `json:"file_resolution"`
	BackupOnWrite  ConfigValue      `json:"backup_on_write"`
}
