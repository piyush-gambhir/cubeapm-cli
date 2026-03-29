package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value from the current profile",
		Long: `Get a configuration value from the current profile.

Reads the value from the resolved configuration (profile defaults applied).
Sensitive values (token, password) are masked for security.

Valid keys:
  server           CubeAPM server address
  email            Login email
  password         Login password (masked)
  auth_method      Authentication method (kratos or none)
  query_port       Query API port
  ingest_port      Ingest API port
  admin_port       Admin API port
  output           Default output format
  current_profile  Name of the currently active profile

Examples:
  # Get the server address
  cubeapm config get server

  # Get the auth method
  cubeapm config get auth_method

  # Get the currently active profile name
  cubeapm config get current_profile`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if key == "current_profile" {
				if cfg.CurrentProfile == "" {
					fmt.Println("(not set)")
				} else {
					fmt.Println(cfg.CurrentProfile)
				}
				return nil
			}

			profile := cfg.GetCurrentProfile().WithDefaults()

			switch key {
			case "server":
				fmt.Println(profile.Server)
			case "query_port":
				fmt.Println(profile.QueryPort)
			case "ingest_port":
				fmt.Println(profile.IngestPort)
			case "admin_port":
				fmt.Println(profile.AdminPort)
			case "email":
				if profile.Email == "" {
					fmt.Println("(not set)")
				} else {
					fmt.Println(profile.Email)
				}
			case "password":
				if profile.Password == "" {
					fmt.Println("(not set)")
				} else {
					fmt.Println("****")
				}
			case "auth_method":
				if profile.AuthMethod == "" {
					fmt.Println("(auto)")
				} else {
					fmt.Println(profile.AuthMethod)
				}
			case "output":
				fmt.Println(profile.Output)
			default:
				return fmt.Errorf("unknown config key %q: valid keys are server, email, password, auth_method, query_port, ingest_port, admin_port, output, current_profile", key)
			}

			return nil
		},
	}
}
