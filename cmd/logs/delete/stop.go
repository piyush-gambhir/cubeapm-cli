package delete

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
)

func newStopCmd() *cobra.Command {
	var ifExists bool

	cmd := &cobra.Command{
		Use:         "stop <task-id>",
		Short:       "Stop a running log deletion task",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Stop a running log deletion task by its task ID.

Sends a stop request to the CubeAPM admin API for the specified deletion task.
The task ID can be obtained from 'cubeapm logs delete list'.

Note: Entries already deleted before the stop request will not be restored.

This command uses the admin API port (default: 3199).

Flags:
  --if-exists  Succeed silently if the task does not exist (idempotent mode)

Examples:
  # Stop a deletion task
  cubeapm logs delete stop abc123-def456

  # Stop idempotently (no error if task doesn't exist)
  cubeapm logs delete stop --if-exists abc123-def456

  # Workflow: list tasks, then stop one
  cubeapm logs delete list
  cubeapm logs delete stop <task-id-from-list>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			if err := cmdutil.APIClient.DeleteLogsStop(taskID); err != nil {
				if ifExists && isNotFoundError(err) {
					if !cmdutil.Quiet {
						fmt.Printf("Delete task %s not found (ignored due to --if-exists).\n", taskID)
					}
					return nil
				}
				return err
			}

			if !cmdutil.Quiet {
				fmt.Printf("Delete task %s stopped.\n", taskID)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&ifExists, "if-exists", false, "Succeed silently if the task does not exist")

	return cmd
}

// isNotFoundError checks if the error indicates a 404 / not-found condition.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "404") || strings.Contains(msg, "not found")
}
