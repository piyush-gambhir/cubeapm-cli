package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List active log deletion tasks",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := cmdutil.APIClient.DeleteLogsList()
			if err != nil {
				return err
			}

			if len(tasks) == 0 {
				fmt.Println("No active deletion tasks.")
				return nil
			}

			table := output.TableDef{
				Headers: []string{"TASK_ID", "FILTER", "STATUS", "PROGRESS"},
			}
			for _, t := range tasks {
				table.Rows = append(table.Rows, []string{
					t.TaskID,
					t.Filter,
					t.Status,
					t.Progress,
				})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}
}
