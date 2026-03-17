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

Examples:
  cubeapm logs hits --query 'error' --last 1h --step 5m
  cubeapm logs hits --query '*' --last 24h --step 1h`,
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
