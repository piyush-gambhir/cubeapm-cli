package config

import (
	"fmt"
	"os"
	"strconv"
)

func ValidateSelectedProfile(cfg *Config, requested string) error {
	name := requested
	if name == "" {
		name = cfg.CurrentProfile
	}
	if name == "" {
		return nil
	}
	if _, ok := cfg.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	return nil
}

type ResolvedConfig struct {
	Server     string
	QueryPort  int
	IngestPort int
	AdminPort  int
	Output     string
	Verbose    bool
	NoColor    bool
	ReadOnly   bool

	AuthMethod    string
	Email         string
	Password      string
	SessionCookie string
	SessionExpiry string
}

type FlagOverrides struct {
	Server     string
	QueryPort  int
	IngestPort int
	AdminPort  int
	Output     string
	Profile    string
	Verbose    bool
	NoColor    bool

	Email    string
	Password string
}

func ResolveAuth(cfg *Config, flags FlagOverrides) ResolvedConfig {
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
		Server:        profile.Server,
		QueryPort:     profile.QueryPort,
		IngestPort:    profile.IngestPort,
		AdminPort:     profile.AdminPort,
		Output:        profile.Output,
		Verbose:       flags.Verbose,
		NoColor:       flags.NoColor,
		ReadOnly:      profile.ReadOnly,
		AuthMethod:    profile.AuthMethod,
		Email:         profile.Email,
		Password:      profile.Password,
		SessionCookie: profile.SessionCookie,
		SessionExpiry: profile.SessionExpiry,
	}

	if v := os.Getenv("CUBEAPM_SERVER"); v != "" {
		resolved.Server = v
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
	if v := os.Getenv("CUBEAPM_READ_ONLY"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil && b {
			resolved.ReadOnly = true
		}
	}
	if v := os.Getenv("CUBEAPM_EMAIL"); v != "" {
		resolved.Email = v
	}
	if v := os.Getenv("CUBEAPM_PASSWORD"); v != "" {
		resolved.Password = v
	}

	if flags.Server != "" {
		resolved.Server = flags.Server
	}
	if flags.Email != "" {
		resolved.Email = flags.Email
	}
	if flags.Password != "" {
		resolved.Password = flags.Password
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
