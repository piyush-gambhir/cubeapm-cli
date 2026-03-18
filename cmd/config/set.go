package config

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value in the current profile",
		Long: `Set a configuration value in the current profile.

Saves the value to the config file (~/.config/cubeapm/config.yaml).

Valid keys:
  server       CubeAPM server address (hostname or IP)
  token        Authentication token for API requests
  query_port   Query API port (default: 3140)
  ingest_port  Ingest API port (default: 3130)
  admin_port   Admin API port (default: 3199)
  output       Default output format: table, json, or yaml

Examples:
  # Set the server address
  cubeapm config set server cubeapm.example.com

  # Set the default output format to JSON
  cubeapm config set output json

  # Set a custom query port
  cubeapm config set query_port 3140

  # Set an authentication token
  cubeapm config set token my-api-token`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			profileName := cfg.CurrentProfile
			if profileName == "" {
				profileName = "default"
				cfg.CurrentProfile = profileName
			}

			profile := cfg.GetCurrentProfile()

			switch key {
			case "server":
				profile.Server = value
			case "token":
				profile.Token = value
			case "query_port":
				port, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("invalid port number: %s", value)
				}
				profile.QueryPort = port
			case "ingest_port":
				port, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("invalid port number: %s", value)
				}
				profile.IngestPort = port
			case "admin_port":
				port, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("invalid port number: %s", value)
				}
				profile.AdminPort = port
			case "output":
				if value != "table" && value != "json" && value != "yaml" {
					return fmt.Errorf("invalid output format %q: use table, json, or yaml", value)
				}
				profile.Output = value
			default:
				return fmt.Errorf("unknown config key %q: valid keys are server, token, query_port, ingest_port, admin_port, output", key)
			}

			cfg.SetProfile(profileName, profile)

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			if !cmdutil.Quiet {
				fmt.Printf("Set %s = %s (profile: %s)\n", key, value, profileName)
			}
			return nil
		},
	}
}
