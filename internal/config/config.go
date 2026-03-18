package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultQueryPort  = 3140
	DefaultIngestPort = 3130
	DefaultAdminPort  = 3199
	DefaultOutput     = "table"
)

// Config represents the CLI configuration file structure.
type Config struct {
	CurrentProfile string             `yaml:"current_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
}

// Profile represents a named connection profile.
type Profile struct {
	Server     string `yaml:"server"`
	QueryPort  int    `yaml:"query_port,omitempty"`
	IngestPort int    `yaml:"ingest_port,omitempty"`
	AdminPort  int    `yaml:"admin_port,omitempty"`
	Token      string `yaml:"token"`
	Output     string `yaml:"output,omitempty"`
	ReadOnly   bool   `yaml:"read_only,omitempty"`
}

// ConfigDir returns the path to the configuration directory.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "cubeapm-cli")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "cubeapm-cli")
}

// ConfigPath returns the full path to the configuration file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// Load reads and parses the configuration file.
// If the file does not exist, it returns a default empty config.
func Load() (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
	}

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return cfg, nil
}

// Save writes the configuration to disk.
func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(ConfigPath(), data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// GetCurrentProfile returns the currently active profile, or an empty Profile
// with default ports if no profile is set.
func (c *Config) GetCurrentProfile() Profile {
	if c.CurrentProfile != "" {
		if p, ok := c.Profiles[c.CurrentProfile]; ok {
			return p
		}
	}
	return Profile{}
}

// WithDefaults returns a copy of the profile with default values applied
// for any unset fields.
func (p Profile) WithDefaults() Profile {
	out := p
	if out.QueryPort == 0 {
		out.QueryPort = DefaultQueryPort
	}
	if out.IngestPort == 0 {
		out.IngestPort = DefaultIngestPort
	}
	if out.AdminPort == 0 {
		out.AdminPort = DefaultAdminPort
	}
	if out.Output == "" {
		out.Output = DefaultOutput
	}
	return out
}
