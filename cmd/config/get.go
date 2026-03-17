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
Token values are masked for security (first 4 and last 4 characters shown).

Valid keys:
  server           CubeAPM server address
  token            Authentication token (masked in output)
  query_port       Query API port
  ingest_port      Ingest API port
  admin_port       Admin API port
  output           Default output format
  current_profile  Name of the currently active profile

Examples:
  # Get the server address
  cubeapm config get server

  # Get the default output format
  cubeapm config get output

  # Get the currently active profile name
  cubeapm config get current_profile

  # Get the query port
  cubeapm config get query_port`,
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
			case "token":
				if profile.Token == "" {
					fmt.Println("(not set)")
				} else {
					// Mask the token for security
					if len(profile.Token) > 8 {
						fmt.Printf("%s...%s\n", profile.Token[:4], profile.Token[len(profile.Token)-4:])
					} else {
						fmt.Println("****")
					}
				}
			case "query_port":
				fmt.Println(profile.QueryPort)
			case "ingest_port":
				fmt.Println(profile.IngestPort)
			case "admin_port":
				fmt.Println(profile.AdminPort)
			case "output":
				fmt.Println(profile.Output)
			default:
				return fmt.Errorf("unknown config key %q: valid keys are server, token, query_port, ingest_port, admin_port, output, current_profile", key)
			}

			return nil
		},
	}
}
