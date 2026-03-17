package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <filter>",
		Short: "Start a log deletion task",
		Long: `Start a log deletion task with the given filter expression.

Examples:
  cubeapm logs delete run '_time:<24h AND service:test'
  cubeapm logs delete run '_stream:{env="staging"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := args[0]

			taskID, err := cmdutil.APIClient.DeleteLogsRun(filter)
			if err != nil {
				return err
			}

			fmt.Printf("Delete task started: %s\n", taskID)
			fmt.Println("Use 'cubeapm logs delete list' to check progress.")
			return nil
		},
	}
}
