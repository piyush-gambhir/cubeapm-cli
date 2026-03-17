package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show the current resolved configuration",
		Long: `Show the full resolved configuration in YAML format.

Displays the configuration file path and the contents of the config file,
including all profiles and the currently active profile.

Examples:
  # View the current configuration
  cubeapm config view`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			fmt.Printf("Config file: %s\n\n", config.ConfigPath())

			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshaling config: %w", err)
			}

			_, err = os.Stdout.Write(data)
			return err
		},
	}
}
