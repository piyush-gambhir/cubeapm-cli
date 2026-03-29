package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show the current resolved configuration",
		Long: `Show the full resolved configuration in YAML format.

Displays the configuration file path and the contents of the config file,
including all profiles and the currently active profile.
Sensitive fields (password, session_cookie) are masked.

Examples:
  # View the current configuration
  cubeapm config view`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if !cmdutil.Quiet {
				fmt.Printf("Config file: %s\n\n", config.ConfigPath())
			}

			// Mask sensitive fields before displaying
			sanitized := sanitizeConfig(cfg)

			data, err := yaml.Marshal(sanitized)
			if err != nil {
				return fmt.Errorf("marshaling config: %w", err)
			}

			_, err = os.Stdout.Write(data)
			return err
		},
	}
}

func sanitizeConfig(cfg *config.Config) *config.Config {
	out := &config.Config{
		CurrentProfile: cfg.CurrentProfile,
		Profiles:       make(map[string]config.Profile, len(cfg.Profiles)),
	}
	for name, p := range cfg.Profiles {
		if p.Password != "" {
			p.Password = "****"
		}
		if p.SessionCookie != "" {
			p.SessionCookie = "****"
		}
		if p.Token != "" && len(p.Token) > 8 {
			p.Token = p.Token[:4] + "..." + p.Token[len(p.Token)-4:]
		} else if p.Token != "" {
			p.Token = "****"
		}
		out.Profiles[name] = p
	}
	return out
}
