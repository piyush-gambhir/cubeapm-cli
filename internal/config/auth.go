package config

import (
	"os"
	"strconv"
)

// ResolvedConfig holds the final resolved configuration after layering
// flags > env vars > config profile.
type ResolvedConfig struct {
	Server     string
	QueryPort  int
	IngestPort int
	AdminPort  int
	Token      string
	Output     string
	Verbose    bool
	NoColor    bool
}

// FlagOverrides holds values from CLI flags that may override config.
type FlagOverrides struct {
	Server     string
	Token      string
	QueryPort  int
	IngestPort int
	AdminPort  int
	Output     string
	Profile    string
	Verbose    bool
	NoColor    bool
}

// ResolveAuth resolves the final configuration by layering:
// flags > environment variables > config profile.
func ResolveAuth(cfg *Config, flags FlagOverrides) ResolvedConfig {
	// Start with config profile
	var profileName string
	if flags.Profile != "" {
		profileName = flags.Profile
	} else {
		profileName = cfg.CurrentProfile
	}

	var profile Profile
	if profileName != "" {
		if p, ok := cfg.Profiles[profileName]; ok {
			profile = p
		}
	}
	profile = profile.WithDefaults()

	resolved := ResolvedConfig{
		Server:     profile.Server,
		QueryPort:  profile.QueryPort,
		IngestPort: profile.IngestPort,
		AdminPort:  profile.AdminPort,
		Token:      profile.Token,
		Output:     profile.Output,
		Verbose:    flags.Verbose,
		NoColor:    flags.NoColor,
	}

	// Layer environment variables
	if v := os.Getenv("CUBEAPM_SERVER"); v != "" {
		resolved.Server = v
	}
	if v := os.Getenv("CUBEAPM_TOKEN"); v != "" {
		resolved.Token = v
	}
	if v := os.Getenv("CUBEAPM_QUERY_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			resolved.QueryPort = port
		}
	}
	if v := os.Getenv("CUBEAPM_INGEST_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			resolved.IngestPort = port
		}
	}
	if v := os.Getenv("CUBEAPM_ADMIN_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			resolved.AdminPort = port
		}
	}

	// Layer flag overrides (highest priority)
	if flags.Server != "" {
		resolved.Server = flags.Server
	}
	if flags.Token != "" {
		resolved.Token = flags.Token
	}
	if flags.QueryPort != 0 {
		resolved.QueryPort = flags.QueryPort
	}
	if flags.IngestPort != 0 {
		resolved.IngestPort = flags.IngestPort
	}
	if flags.AdminPort != 0 {
		resolved.AdminPort = flags.AdminPort
	}
	if flags.Output != "" {
		resolved.Output = flags.Output
	}

	// Ensure defaults
	if resolved.QueryPort == 0 {
		resolved.QueryPort = DefaultQueryPort
	}
	if resolved.IngestPort == 0 {
		resolved.IngestPort = DefaultIngestPort
	}
	if resolved.AdminPort == 0 {
		resolved.AdminPort = DefaultAdminPort
	}
	if resolved.Output == "" {
		resolved.Output = DefaultOutput
	}

	return resolved
}
