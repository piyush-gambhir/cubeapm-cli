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

Runs a LogsQL stats pipeline that performs aggregations over matching log
entries. The query must contain a '| stats' pipe that defines what to compute.

Common stats functions:
  count()          - count matching entries
  count_uniq(f)    - count unique values of field f
  sum(f)           - sum of numeric field f
  avg(f)           - average of numeric field f
  min(f), max(f)   - minimum/maximum of field f
  median(f)        - median of field f
  quantile(0.99,f) - percentile of field f
  values(f)        - list of unique values

Stats queries follow the pattern: <filter> | stats <function> by (<fields>)

Time ranges can be specified as:
  - Relative:   --last 1h  (also: 30m, 2d)
  - RFC3339:    --from 2024-01-15T00:00:00Z
  - Default:    last 1 hour if no time flags are provided

Examples:
  # Count log entries by status in the last hour
  cubeapm logs stats '_time:1h | stats count() by (status)'

  # Count errors by service in the last 24 hours
  cubeapm logs stats 'error | stats count() by (service)' --last 24h

  # Count unique users per service
  cubeapm logs stats '* | stats count_uniq(user_id) by (service)' --last 1h

  # Get top error messages
  cubeapm logs stats 'level:error | stats count() by (_msg)' --last 1h

  # Output as JSON
  cubeapm logs stats 'error | stats count() by (service)' --last 24h -o json`,
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
