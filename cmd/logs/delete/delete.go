package delete

import (
	"github.com/spf13/cobra"
)

// NewDeleteCmd creates the "delete" parent command under logs.
func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Manage log deletion tasks",
		Long: `Start, stop, and list log deletion tasks via the admin API.

Log deletion is an asynchronous operation. You submit a deletion task with
a filter expression, and CubeAPM processes it in the background. Use the
subcommands to manage these tasks.

Subcommands:
  run    Start a new deletion task with a filter expression
  stop   Stop a running deletion task by its task ID
  list   List all active (in-progress) deletion tasks

Deletion tasks use the admin API port (default: 3199).

Examples:
  cubeapm logs delete run '_time:<24h AND service:test'
  cubeapm logs delete list
  cubeapm logs delete stop <task-id>`,
	}

	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}
