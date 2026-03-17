package delete

import (
	"github.com/spf13/cobra"
)

// NewDeleteCmd creates the "delete" parent command under logs.
func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Manage log deletion tasks",
		Long:  "Start, stop, and list log deletion tasks via the admin API.",
	}

	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}
