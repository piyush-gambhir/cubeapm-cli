package logs

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newHitsCmd() *cobra.Command {
	var (
		query string
		from  string
		to    string
		last  string
		step  string
	)

	cmd := &cobra.Command{
		Use:   "hits",
		Short: "Show log volume over time",
		Long: `Show the number of log entries matching a query over time buckets.

Queries the VictoriaLogs-compatible hits endpoint to return a histogram
of log entry counts over time. This is useful for understanding log volume
patterns, identifying spikes, or visualizing error rates.

The --query flag accepts a LogsQL expression (default: '*' to match all logs).
The --step flag controls the time bucket size (e.g., 5m, 1h). If not set,
the server auto-selects an appropriate bucket size.

Time ranges can be specified as:
  - Relative:   --last 1h  (also: 30m, 2d)
  - RFC3339:    --from 2024-01-15T00:00:00Z --to 2024-01-15T12:00:00Z
  - Default:    last 1 hour if no time flags are provided

Examples:
  # Show all log volume over the last hour in 5-minute buckets
  cubeapm logs hits --query '*' --last 1h --step 5m

  # Show error log volume over the last 24 hours in 1-hour buckets
  cubeapm logs hits --query 'error' --last 24h --step 1h

  # Show volume for a specific service
  cubeapm logs hits --query 'service:api-gateway' --last 6h --step 15m

  # Output as JSON for graphing
  cubeapm logs hits --query 'error' --last 24h --step 1h -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if query == "" {
				query = "*"
			}

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			result, err := cmdutil.APIClient.GetLogHits(query, start, end, step)
			if err != nil {
				return err
			}

			if len(result.Hits) == 0 {
				fmt.Println("No log hits data returned.")
				return nil
			}

			table := output.TableDef{
				Headers: []string{"TIMESTAMP", "COUNT"},
			}

			for _, hit := range result.Hits {
				for i := range hit.Timestamps {
					count := int64(0)
					if i < len(hit.Values) {
						count = hit.Values[i]
					}
					table.Rows = append(table.Rows, []string{
						hit.Timestamps[i],
						strconv.FormatInt(count, 10),
					})
				}
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "LogsQL query (default: *)")
	cmd.Flags().StringVar(&step, "step", "", "Time bucket step (e.g., 5m, 1h)")
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
