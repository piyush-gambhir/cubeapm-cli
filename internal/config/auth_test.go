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
				Email:      "user@example.com",
				Password:   "secret",
				AuthMethod: "kratos",
				Output:     "json",
			},
		},
	}

	resolved := ResolveAuth(cfg, FlagOverrides{})

	if resolved.Server != "cubeapm.example.com" {
		t.Errorf("Server = %q, want %q", resolved.Server, "cubeapm.example.com")
	}
	if resolved.Email != "user@example.com" {
		t.Errorf("Email = %q, want %q", resolved.Email, "user@example.com")
	}
	if resolved.Password != "secret" {
		t.Errorf("Password = %q, want %q", resolved.Password, "secret")
	}
	if resolved.AuthMethod != "kratos" {
		t.Errorf("AuthMethod = %q, want %q", resolved.AuthMethod, "kratos")
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
				Email:  "config@example.com",
			},
		},
	}

	t.Setenv("CUBEAPM_SERVER", "env-server.com")
	t.Setenv("CUBEAPM_EMAIL", "env@example.com")
	t.Setenv("CUBEAPM_PASSWORD", "env-pass")
	t.Setenv("CUBEAPM_QUERY_PORT", "4140")
	t.Setenv("CUBEAPM_INGEST_PORT", "4130")
	t.Setenv("CUBEAPM_ADMIN_PORT", "4199")

	resolved := ResolveAuth(cfg, FlagOverrides{})

	if resolved.Server != "env-server.com" {
		t.Errorf("Server = %q, want %q (env should override config)", resolved.Server, "env-server.com")
	}
	if resolved.Email != "env@example.com" {
		t.Errorf("Email = %q, want %q (env should override config)", resolved.Email, "env@example.com")
	}
	if resolved.Password != "env-pass" {
		t.Errorf("Password = %q, want %q (env should override config)", resolved.Password, "env-pass")
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
				Email:  "config@example.com",
			},
		},
	}

	t.Setenv("CUBEAPM_SERVER", "env-server.com")
	t.Setenv("CUBEAPM_EMAIL", "env@example.com")

	flags := FlagOverrides{
		Server:     "flag-server.com",
		Email:      "flag@example.com",
		Password:   "flag-pass",
		QueryPort:  5140,
		IngestPort: 5130,
		AdminPort:  5199,
		Output:     "yaml",
	}

	resolved := ResolveAuth(cfg, flags)

	if resolved.Server != "flag-server.com" {
		t.Errorf("Server = %q, want %q (flags should override env)", resolved.Server, "flag-server.com")
	}
	if resolved.Email != "flag@example.com" {
		t.Errorf("Email = %q, want %q (flags should override env)", resolved.Email, "flag@example.com")
	}
	if resolved.Password != "flag-pass" {
		t.Errorf("Password = %q, want %q (flags should override env)", resolved.Password, "flag-pass")
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
				Email:      "admin@example.com",
				Password:   "pass",
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
