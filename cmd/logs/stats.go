package logs

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newStatsCmd() *cobra.Command {
	var (
		from string
		to   string
		last string
	)

	cmd := &cobra.Command{
		Use:   "stats <logsql>",
		Short: "Execute a stats query on logs",
		Long: `Execute a LogsQL stats query and display aggregated results.

Examples:
  cubeapm logs stats '_time:1h | stats count() by (status)'
  cubeapm logs stats 'error | stats count() by (service)' --last 24h`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logsql := args[0]

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			result, err := cmdutil.APIClient.GetLogStats(logsql, start, end)
			if err != nil {
				return err
			}

			if len(result.Rows) == 0 {
				fmt.Println("No stats data returned.")
				return nil
			}

			// Discover all field names from rows
			fieldSet := make(map[string]bool)
			for _, row := range result.Rows {
				for k := range row.Fields {
					fieldSet[k] = true
				}
			}

			headers := make([]string, 0, len(fieldSet))
			for k := range fieldSet {
				headers = append(headers, k)
			}
			sort.Strings(headers)

			table := output.TableDef{
				Headers: headers,
			}
			for _, row := range result.Rows {
				cells := make([]string, len(headers))
				for i, h := range headers {
					cells[i] = row.Fields[h]
				}
				table.Rows = append(table.Rows, cells)
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
