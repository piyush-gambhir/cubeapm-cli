package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadConfig_NewFile(t *testing.T) {
	// Point config dir at a temporary directory with no config file
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.Profiles == nil {
		t.Fatal("Load() Profiles map is nil, expected initialized map")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("got %d profiles, want 0", len(cfg.Profiles))
	}
	if cfg.CurrentProfile != "" {
		t.Errorf("CurrentProfile = %q, want empty", cfg.CurrentProfile)
	}
}

func TestLoadConfig_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Write a config file manually
	configDir := filepath.Join(tmpDir, "cubeapm-cli")
	os.MkdirAll(configDir, 0700)
	configData := Config{
		CurrentProfile: "production",
		Profiles: map[string]Profile{
			"production": {
				Server:     "cubeapm.example.com",
				QueryPort:  3140,
				IngestPort: 3130,
				AdminPort:  3199,
				Token:      "prod-token",
			},
		},
	}
	data, _ := yaml.Marshal(configData)
	os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.CurrentProfile != "production" {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, "production")
	}
	if len(cfg.Profiles) != 1 {
		t.Errorf("got %d profiles, want 1", len(cfg.Profiles))
	}
	p := cfg.Profiles["production"]
	if p.Server != "cubeapm.example.com" {
		t.Errorf("Server = %q, want %q", p.Server, "cubeapm.example.com")
	}
	if p.Token != "prod-token" {
		t.Errorf("Token = %q, want %q", p.Token, "prod-token")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := &Config{
		CurrentProfile: "dev",
		Profiles: map[string]Profile{
			"dev": {
				Server:     "localhost",
				QueryPort:  3140,
				IngestPort: 3130,
				AdminPort:  3199,
				Token:      "dev-token",
			},
		},
	}

	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify the file was created
	configPath := filepath.Join(tmpDir, "cubeapm-cli", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading saved config: %v", err)
	}

	var loaded Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("parsing saved config: %v", err)
	}

	if loaded.CurrentProfile != "dev" {
		t.Errorf("saved CurrentProfile = %q, want %q", loaded.CurrentProfile, "dev")
	}
	if loaded.Profiles["dev"].Server != "localhost" {
		t.Errorf("saved Server = %q, want %q", loaded.Profiles["dev"].Server, "localhost")
	}
}

func TestWithDefaults(t *testing.T) {
	// Profile with no ports set
	p := Profile{
		Server: "example.com",
		Token:  "tok",
	}

	result := p.WithDefaults()

	if result.QueryPort != DefaultQueryPort {
		t.Errorf("QueryPort = %d, want %d", result.QueryPort, DefaultQueryPort)
	}
	if result.IngestPort != DefaultIngestPort {
		t.Errorf("IngestPort = %d, want %d", result.IngestPort, DefaultIngestPort)
	}
	if result.AdminPort != DefaultAdminPort {
		t.Errorf("AdminPort = %d, want %d", result.AdminPort, DefaultAdminPort)
	}
	if result.Output != DefaultOutput {
		t.Errorf("Output = %q, want %q", result.Output, DefaultOutput)
	}

	// Profile with custom ports should keep them
	p2 := Profile{
		Server:     "example.com",
		QueryPort:  9090,
		IngestPort: 9091,
		AdminPort:  9092,
		Output:     "json",
	}

	result2 := p2.WithDefaults()
	if result2.QueryPort != 9090 {
		t.Errorf("QueryPort = %d, want 9090", result2.QueryPort)
	}
	if result2.IngestPort != 9091 {
		t.Errorf("IngestPort = %d, want 9091", result2.IngestPort)
	}
	if result2.AdminPort != 9092 {
		t.Errorf("AdminPort = %d, want 9092", result2.AdminPort)
	}
	if result2.Output != "json" {
		t.Errorf("Output = %q, want %q", result2.Output, "json")
	}
}

func TestGetCurrentProfile(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "prod",
		Profiles: map[string]Profile{
			"prod": {
				Server: "prod.example.com",
				Token:  "prod-token",
			},
			"dev": {
				Server: "localhost",
				Token:  "dev-token",
			},
		},
	}

	p := cfg.GetCurrentProfile()
	if p.Server != "prod.example.com" {
		t.Errorf("Server = %q, want %q", p.Server, "prod.example.com")
	}

	// Test with no current profile set
	cfg.CurrentProfile = ""
	p = cfg.GetCurrentProfile()
	if p.Server != "" {
		t.Errorf("expected empty profile for no current profile, got Server = %q", p.Server)
	}

	// Test with non-existent current profile
	cfg.CurrentProfile = "nonexistent"
	p = cfg.GetCurrentProfile()
	if p.Server != "" {
		t.Errorf("expected empty profile for nonexistent current profile, got Server = %q", p.Server)
	}
}

func TestConfigDir(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	dir := ConfigDir()
	expected := filepath.Join(tmpDir, "cubeapm-cli")
	if dir != expected {
		t.Errorf("ConfigDir() = %q, want %q", dir, expected)
	}
}
