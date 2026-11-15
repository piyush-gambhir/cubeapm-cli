package config

import "fmt"

func (c *Config) CreateProfile(name string, p Profile) error {
	if _, exists := c.Profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}
	c.Profiles[name] = p
	return nil
}

func (c *Config) UpdateProfile(name string, p Profile) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q does not exist", name)
	}
	c.Profiles[name] = p
	return nil
}

func (c *Config) SetProfile(name string, p Profile) {
	c.Profiles[name] = p
}

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

func (c *Config) SetCurrentProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q does not exist", name)
	}
	c.CurrentProfile = name
	return nil
}

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

func (c *Config) ListProfiles() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	return names
}
