package logs

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newStreamsCmd() *cobra.Command {
	var (
		query string
		from  string
		to    string
		last  string
	)

	cmd := &cobra.Command{
		Use:   "streams",
		Short: "List log streams",
		Long: `List log streams and their entry counts.

A log stream in VictoriaLogs is a unique combination of stream-level labels
(e.g., host, container, pod). Each stream groups related log entries together.

This command returns all streams and the number of log entries in each stream,
optionally filtered by a LogsQL query.

Time ranges can be specified as:
  - Relative:   --last 1h  (also: 30m, 2d)
  - RFC3339:    --from 2024-01-15T00:00:00Z
  - Default:    last 1 hour if no time flags are provided

Examples:
  # List all streams in the last hour
  cubeapm logs streams --last 1h

  # List streams containing errors in the last 24 hours
  cubeapm logs streams --query 'error' --last 24h

  # List streams for a specific service
  cubeapm logs streams --query 'service:api-gateway' --last 1h

  # Output as JSON
  cubeapm logs streams --last 1h -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			streams, err := cmdutil.APIClient.GetLogStreams(query, start, end)
			if err != nil {
				return err
			}

			if len(streams) == 0 {
				fmt.Println("No streams found.")
				return nil
			}

			table := output.TableDef{
				Headers: []string{"STREAM", "ENTRIES"},
			}
			for _, s := range streams {
				table.Rows = append(table.Rows, []string{
					s.Stream,
					strconv.FormatInt(s.Entries, 10),
				})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "LogsQL query to filter streams")
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
