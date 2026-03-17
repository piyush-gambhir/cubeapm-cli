package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <task-id>",
		Short: "Stop a running log deletion task",
		Long: `Stop a running log deletion task by its task ID.

Sends a stop request to the CubeAPM admin API for the specified deletion task.
The task ID can be obtained from 'cubeapm logs delete list'.

Note: Entries already deleted before the stop request will not be restored.

This command uses the admin API port (default: 3199).

Examples:
  # Stop a deletion task
  cubeapm logs delete stop abc123-def456

  # Workflow: list tasks, then stop one
  cubeapm logs delete list
  cubeapm logs delete stop <task-id-from-list>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			if err := cmdutil.APIClient.DeleteLogsStop(taskID); err != nil {
				return err
			}

			fmt.Printf("Delete task %s stopped.\n", taskID)
			return nil
		},
	}
}
