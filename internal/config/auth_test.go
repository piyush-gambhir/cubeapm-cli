package config

import (
	"testing"
)

func TestResolveAuth_ConfigOnly(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "prod",
		Profiles: map[string]Profile{
			"prod": {
				Server:     "cubeapm.example.com",
				QueryPort:  3140,
				IngestPort: 3130,
				AdminPort:  3199,
				Token:      "prod-token",
				Output:     "json",
			},
		},
	}

	resolved := ResolveAuth(cfg, FlagOverrides{})

	if resolved.Server != "cubeapm.example.com" {
		t.Errorf("Server = %q, want %q", resolved.Server, "cubeapm.example.com")
	}
	if resolved.Token != "prod-token" {
		t.Errorf("Token = %q, want %q", resolved.Token, "prod-token")
	}
	if resolved.QueryPort != 3140 {
		t.Errorf("QueryPort = %d, want %d", resolved.QueryPort, 3140)
	}
	if resolved.IngestPort != 3130 {
		t.Errorf("IngestPort = %d, want %d", resolved.IngestPort, 3130)
	}
	if resolved.AdminPort != 3199 {
		t.Errorf("AdminPort = %d, want %d", resolved.AdminPort, 3199)
	}
	if resolved.Output != "json" {
		t.Errorf("Output = %q, want %q", resolved.Output, "json")
	}
}

func TestResolveAuth_EnvOverrides(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "prod",
		Profiles: map[string]Profile{
			"prod": {
				Server: "config-server.com",
				Token:  "config-token",
			},
		},
	}

	t.Setenv("CUBEAPM_SERVER", "env-server.com")
	t.Setenv("CUBEAPM_TOKEN", "env-token")
	t.Setenv("CUBEAPM_QUERY_PORT", "4140")
	t.Setenv("CUBEAPM_INGEST_PORT", "4130")
	t.Setenv("CUBEAPM_ADMIN_PORT", "4199")

	resolved := ResolveAuth(cfg, FlagOverrides{})

	if resolved.Server != "env-server.com" {
		t.Errorf("Server = %q, want %q (env should override config)", resolved.Server, "env-server.com")
	}
	if resolved.Token != "env-token" {
		t.Errorf("Token = %q, want %q (env should override config)", resolved.Token, "env-token")
	}
	if resolved.QueryPort != 4140 {
		t.Errorf("QueryPort = %d, want %d (env should override config)", resolved.QueryPort, 4140)
	}
	if resolved.IngestPort != 4130 {
		t.Errorf("IngestPort = %d, want %d (env should override config)", resolved.IngestPort, 4130)
	}
	if resolved.AdminPort != 4199 {
		t.Errorf("AdminPort = %d, want %d (env should override config)", resolved.AdminPort, 4199)
	}
}

func TestResolveAuth_FlagOverrides(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "prod",
		Profiles: map[string]Profile{
			"prod": {
				Server: "config-server.com",
				Token:  "config-token",
			},
		},
	}

	t.Setenv("CUBEAPM_SERVER", "env-server.com")
	t.Setenv("CUBEAPM_TOKEN", "env-token")

	flags := FlagOverrides{
		Server:     "flag-server.com",
		Token:      "flag-token",
		QueryPort:  5140,
		IngestPort: 5130,
		AdminPort:  5199,
		Output:     "yaml",
	}

	resolved := ResolveAuth(cfg, flags)

	if resolved.Server != "flag-server.com" {
		t.Errorf("Server = %q, want %q (flags should override env)", resolved.Server, "flag-server.com")
	}
	if resolved.Token != "flag-token" {
		t.Errorf("Token = %q, want %q (flags should override env)", resolved.Token, "flag-token")
	}
	if resolved.QueryPort != 5140 {
		t.Errorf("QueryPort = %d, want %d (flags should override env)", resolved.QueryPort, 5140)
	}
	if resolved.IngestPort != 5130 {
		t.Errorf("IngestPort = %d, want %d (flags should override env)", resolved.IngestPort, 5130)
	}
	if resolved.AdminPort != 5199 {
		t.Errorf("AdminPort = %d, want %d (flags should override env)", resolved.AdminPort, 5199)
	}
	if resolved.Output != "yaml" {
		t.Errorf("Output = %q, want %q", resolved.Output, "yaml")
	}
}

func TestResolveAuth_DefaultPorts(t *testing.T) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
	}

	resolved := ResolveAuth(cfg, FlagOverrides{})

	if resolved.QueryPort != DefaultQueryPort {
		t.Errorf("QueryPort = %d, want %d", resolved.QueryPort, DefaultQueryPort)
	}
	if resolved.IngestPort != DefaultIngestPort {
		t.Errorf("IngestPort = %d, want %d", resolved.IngestPort, DefaultIngestPort)
	}
	if resolved.AdminPort != DefaultAdminPort {
		t.Errorf("AdminPort = %d, want %d", resolved.AdminPort, DefaultAdminPort)
	}
	if resolved.Output != DefaultOutput {
		t.Errorf("Output = %q, want %q", resolved.Output, DefaultOutput)
	}
}

func TestResolveAuth_CustomPorts(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "custom",
		Profiles: map[string]Profile{
			"custom": {
				Server:     "custom.example.com",
				QueryPort:  8080,
				IngestPort: 8081,
				AdminPort:  8082,
				Token:      "custom-token",
			},
		},
	}

	resolved := ResolveAuth(cfg, FlagOverrides{})

	if resolved.QueryPort != 8080 {
		t.Errorf("QueryPort = %d, want %d", resolved.QueryPort, 8080)
	}
	if resolved.IngestPort != 8081 {
		t.Errorf("IngestPort = %d, want %d", resolved.IngestPort, 8081)
	}
	if resolved.AdminPort != 8082 {
		t.Errorf("AdminPort = %d, want %d", resolved.AdminPort, 8082)
	}
}
