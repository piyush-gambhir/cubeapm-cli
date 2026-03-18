package config

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

func newProfilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profiles",
		Short: "Manage connection profiles",
		Long: `Manage connection profiles for different CubeAPM instances.

Profiles store server address, authentication token, and port settings.
You can have multiple profiles (e.g., "production", "staging", "local")
and switch between them.

Profiles are created via 'cubeapm login' or 'cubeapm config set'.

Subcommands:
  list    List all profiles (active profile marked with *)
  use     Set the active profile
  delete  Delete a profile

Examples:
  cubeapm config profiles list
  cubeapm config profiles use production
  cubeapm config profiles delete staging`,
	}

	cmd.AddCommand(newProfilesListCmd())
	cmd.AddCommand(newProfilesUseCmd())
	cmd.AddCommand(newProfilesDeleteCmd())

	return cmd
}

func newProfilesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all profiles",
		Long: `List all configured connection profiles.

The currently active profile is marked with an asterisk (*).
Each profile shows its name and the associated server address.

Examples:
  # List all profiles
  cubeapm config profiles list

  # Use the short alias
  cubeapm config profiles ls`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			names := cfg.ListProfiles()
			sort.Strings(names)

			if len(names) == 0 {
				fmt.Println("No profiles configured. Run 'cubeapm login' to create one.")
				return nil
			}

			for _, name := range names {
				profile := cfg.Profiles[name]
				marker := "  "
				if name == cfg.CurrentProfile {
					marker = "* "
				}
				fmt.Printf("%s%-20s %s\n", marker, name, profile.Server)
			}

			return nil
		},
	}
}

func newProfilesUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Set the active profile",
		Long: `Set the active connection profile.

All subsequent commands will use the settings (server, token, ports)
from the selected profile unless overridden by flags or environment variables.

Examples:
  # Switch to the production profile
  cubeapm config profiles use production

  # Switch to the staging profile
  cubeapm config profiles use staging`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if err := cfg.SetCurrentProfile(name); err != nil {
				return err
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			if !cmdutil.Quiet {
				fmt.Printf("Switched to profile %q\n", name)
			}
			return nil
		},
	}
}

func newProfilesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <profile>",
		Short: "Delete a profile",
		Long: `Delete a connection profile.

Removes the specified profile from the config file. If the deleted profile
was the active profile, you will need to select a new one with
'cubeapm config profiles use'.

Examples:
  # Delete a profile
  cubeapm config profiles delete staging

  # Use the short alias
  cubeapm config profiles rm old-profile`,
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if err := cfg.DeleteProfile(name); err != nil {
				return err
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			if !cmdutil.Quiet {
				fmt.Printf("Deleted profile %q\n", name)
			}
			return nil
		},
	}
}
