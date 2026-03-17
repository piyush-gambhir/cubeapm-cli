package config

import (
	"github.com/spf13/cobra"
)

// NewConfigCmd creates the "config" parent command.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long: `View and modify CubeAPM CLI configuration settings and profiles.

Configuration is stored in ~/.config/cubeapm/config.yaml and supports
multiple profiles for connecting to different CubeAPM instances.

Settings can also be overridden via environment variables:
  CUBEAPM_SERVER       Server address
  CUBEAPM_TOKEN        Authentication token
  CUBEAPM_QUERY_PORT   Query port (default: 3140)
  CUBEAPM_INGEST_PORT  Ingest port (default: 3130)
  CUBEAPM_ADMIN_PORT   Admin port (default: 3199)

Or via global CLI flags: --server, --token, --query-port, --ingest-port, --admin-port.

Priority (highest to lowest): CLI flags > environment variables > profile config.

Subcommands:
  view      Show the full resolved configuration
  set       Set a configuration value in the current profile
  get       Get a configuration value from the current profile
  profiles  Manage connection profiles (list, use, delete)

Examples:
  cubeapm config view
  cubeapm config set server cubeapm.example.com
  cubeapm config get server
  cubeapm config profiles list`,
	}

	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newProfilesCmd())

	return cmd
}
