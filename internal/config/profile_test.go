package config

import (
	"sort"
	"testing"
)

func TestCreateProfile(t *testing.T) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
	}

	p := Profile{Server: "example.com", Token: "tok"}
	err := cfg.CreateProfile("production", p)
	if err != nil {
		t.Fatalf("CreateProfile() error = %v", err)
	}

	if _, ok := cfg.Profiles["production"]; !ok {
		t.Fatal("profile 'production' not found after creation")
	}
	if cfg.Profiles["production"].Server != "example.com" {
		t.Errorf("Server = %q, want %q", cfg.Profiles["production"].Server, "example.com")
	}

	// Creating a profile that already exists should fail
	err = cfg.CreateProfile("production", p)
	if err == nil {
		t.Fatal("CreateProfile() expected error for duplicate profile, got nil")
	}
}

func TestUpdateProfile(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]Profile{
			"dev": {Server: "localhost", Token: "old-token"},
		},
	}

	err := cfg.UpdateProfile("dev", Profile{Server: "localhost", Token: "new-token"})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if cfg.Profiles["dev"].Token != "new-token" {
		t.Errorf("Token = %q, want %q", cfg.Profiles["dev"].Token, "new-token")
	}

	// Updating a profile that doesn't exist should fail
	err = cfg.UpdateProfile("nonexistent", Profile{})
	if err == nil {
		t.Fatal("UpdateProfile() expected error for nonexistent profile, got nil")
	}
}

func TestDeleteProfile(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "staging",
		Profiles: map[string]Profile{
			"production": {Server: "prod.example.com"},
			"staging":    {Server: "staging.example.com"},
		},
	}

	// Delete a non-current profile
	err := cfg.DeleteProfile("production")
	if err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
	}
	if _, ok := cfg.Profiles["production"]; ok {
		t.Error("profile 'production' still exists after deletion")
	}
	if cfg.CurrentProfile != "staging" {
		t.Errorf("CurrentProfile = %q, should remain %q", cfg.CurrentProfile, "staging")
	}

	// Delete the current profile should clear CurrentProfile
	err = cfg.DeleteProfile("staging")
	if err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
	}
	if cfg.CurrentProfile != "" {
		t.Errorf("CurrentProfile = %q, want empty after deleting current profile", cfg.CurrentProfile)
	}

	// Deleting a nonexistent profile should fail
	err = cfg.DeleteProfile("nonexistent")
	if err == nil {
		t.Fatal("DeleteProfile() expected error for nonexistent profile, got nil")
	}
}

func TestSetCurrentProfile(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]Profile{
			"production": {Server: "prod.example.com"},
			"staging":    {Server: "staging.example.com"},
		},
	}

	err := cfg.SetCurrentProfile("production")
	if err != nil {
		t.Fatalf("SetCurrentProfile() error = %v", err)
	}
	if cfg.CurrentProfile != "production" {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, "production")
	}

	// Switch to another profile
	err = cfg.SetCurrentProfile("staging")
	if err != nil {
		t.Fatalf("SetCurrentProfile() error = %v", err)
	}
	if cfg.CurrentProfile != "staging" {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, "staging")
	}

	// Setting a nonexistent profile should fail
	err = cfg.SetCurrentProfile("nonexistent")
	if err == nil {
		t.Fatal("SetCurrentProfile() expected error for nonexistent profile, got nil")
	}
}

func TestListProfiles(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]Profile{
			"production": {Server: "prod.example.com"},
			"staging":    {Server: "staging.example.com"},
			"dev":        {Server: "localhost"},
		},
	}

	names := cfg.ListProfiles()
	if len(names) != 3 {
		t.Fatalf("got %d profiles, want 3", len(names))
	}

	sort.Strings(names)
	expected := []string{"dev", "production", "staging"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("names[%d] = %q, want %q", i, name, expected[i])
		}
	}

	// Empty config
	emptyCfg := &Config{Profiles: make(map[string]Profile)}
	emptyNames := emptyCfg.ListProfiles()
	if len(emptyNames) != 0 {
		t.Errorf("got %d profiles for empty config, want 0", len(emptyNames))
	}
}
