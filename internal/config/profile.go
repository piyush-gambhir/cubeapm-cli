package config

import "fmt"

// CreateProfile adds a new profile to the config. Returns an error if the
// profile name already exists.
func (c *Config) CreateProfile(name string, p Profile) error {
	if _, exists := c.Profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}
	c.Profiles[name] = p
	return nil
}

// UpdateProfile updates an existing profile. Returns an error if the profile
// does not exist.
func (c *Config) UpdateProfile(name string, p Profile) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q does not exist", name)
	}
	c.Profiles[name] = p
	return nil
}

// SetProfile creates or updates a profile.
func (c *Config) SetProfile(name string, p Profile) {
	c.Profiles[name] = p
}

// DeleteProfile removes a profile. If the deleted profile was the current
// profile, current_profile is cleared.
func (c *Config) DeleteProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q does not exist", name)
	}
	delete(c.Profiles, name)
	if c.CurrentProfile == name {
		c.CurrentProfile = ""
	}
	return nil
}

// SetCurrentProfile sets the active profile. Returns an error if the profile
// does not exist.
func (c *Config) SetCurrentProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q does not exist", name)
	}
	c.CurrentProfile = name
	return nil
}

// GetActiveProfile returns the currently active profile name and profile.
// Returns an error if no profile is set or the profile doesn't exist.
func (c *Config) GetActiveProfile() (string, Profile, error) {
	if c.CurrentProfile == "" {
		return "", Profile{}, fmt.Errorf("no active profile set; run 'cubeapm login' or 'cubeapm config profiles use <name>'")
	}
	p, ok := c.Profiles[c.CurrentProfile]
	if !ok {
		return "", Profile{}, fmt.Errorf("current profile %q not found in config", c.CurrentProfile)
	}
	return c.CurrentProfile, p, nil
}

// ListProfiles returns all profile names.
func (c *Config) ListProfiles() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	return names
}
