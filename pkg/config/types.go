package config

// Config represents the envx configuration
type Config struct {
	Keystore string `yaml:"keystore,omitempty"`
	File     string `yaml:"file,omitempty"`
	Format   string `yaml:"format,omitempty"`
	KeyName  string `yaml:"key_name,omitempty"`
}

// DefaultConfig returns the default configuration values
func DefaultConfig() *Config {
	return &Config{
		Keystore: "macos",
		File:     ".env",
		Format:   "env",
		KeyName:  "default",
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
	if other.Format != "" {
		c.Format = other.Format
	}
	if other.KeyName != "" {
		c.KeyName = other.KeyName
	}
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

// ConfigReport represents the effective configuration with sources
type ConfigReport struct {
	Keystore ConfigValue `json:"keystore"`
	File     ConfigValue `json:"file"`
	Format   ConfigValue `json:"format"`
	KeyName  ConfigValue `json:"key_name"`
}
