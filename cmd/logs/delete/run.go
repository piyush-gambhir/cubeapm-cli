package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "run <filter>",
		Short:       "Start a log deletion task",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Start a log deletion task with the given filter expression.

Submits a deletion task to the CubeAPM admin API. The task runs asynchronously
in the background. The filter expression uses LogsQL syntax to specify which
log entries to delete.

WARNING: Deletion is irreversible. Make sure your filter expression is correct
before running. Use 'cubeapm logs query' with the same filter to preview
which entries would be affected.

The command returns a task ID that can be used to check progress with
'cubeapm logs delete list' or to stop the task with 'cubeapm logs delete stop'.

This command uses the admin API port (default: 3199).

Examples:
  # Delete logs older than 24 hours for a test service
  cubeapm logs delete run '_time:<24h AND service:test'

  # Delete logs from a specific stream
  cubeapm logs delete run '_stream:{env="staging"}'

  # Delete all debug-level logs older than 7 days
  cubeapm logs delete run '_time:<7d AND level:debug'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := args[0]

			taskID, err := cmdutil.APIClient.DeleteLogsRun(filter)
			if err != nil {
				return err
			}

			if !cmdutil.Quiet {
				fmt.Printf("Delete task started: %s\n", taskID)
				fmt.Println("Use 'cubeapm logs delete list' to check progress.")
			}
			return nil
		},
	}
}
