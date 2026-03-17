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
		Args:  cobra.ExactArgs(1),
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
