package config

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

func newProfilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profiles",
		Short: "Manage connection profiles",
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
		Aliases: []string{"ls"},
		Args:  cobra.NoArgs,
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
		Args:  cobra.ExactArgs(1),
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

			fmt.Printf("Switched to profile %q\n", name)
			return nil
		},
	}
}

func newProfilesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <profile>",
		Short: "Delete a profile",
		Aliases: []string{"rm"},
		Args:  cobra.ExactArgs(1),
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

			fmt.Printf("Deleted profile %q\n", name)
			return nil
		},
	}
}
