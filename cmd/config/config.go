package config

import (
	"github.com/spf13/cobra"
)

// NewConfigCmd creates the "config" parent command.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long:  "View and modify CubeAPM CLI configuration settings and profiles.",
	}

	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newProfilesCmd())

	return cmd
}
