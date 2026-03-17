package logs

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

func newQueryCmd() *cobra.Command {
	var (
		from    string
		to      string
		last    string
		limit   int
		service string
		level   string
		stream  string
	)

	cmd := &cobra.Command{
		Use:   "query <logsql>",
		Short: "Query logs using LogsQL",
		Long: `Query logs using LogsQL syntax (VictoriaLogs-compatible).

Executes a LogsQL query and returns matching log entries. In table mode,
results are collected and displayed as a table. In JSON/YAML mode, results
are streamed line-by-line (newline-delimited).

The <logsql> argument is a LogsQL expression. Common LogsQL syntax includes:
  - Keyword search:    'error'
  - Field filter:      'service:api-gateway'
  - AND/OR/NOT:        'error AND service:api'
  - Stream filter:     '_stream:{host="web-1"}'
  - Regex:             're("pattern")'
  - Range filter:      '_time:1h'

The --service, --level, and --stream flags are convenience shortcuts that
prepend additional filters to the LogsQL query:
  --service api    =>  'service:api AND <logsql>'
  --level error    =>  'level:error AND <logsql>'
  --stream '{host="web-1"}'  =>  '_stream:{host="web-1"} AND <logsql>'

Time ranges can be specified as:
  - Relative:   --last 1h  (also: 30m, 2d, 1d12h)
  - RFC3339:    --from 2024-01-15T10:00:00Z --to 2024-01-15T12:00:00Z
  - Unix:       --from 1705312800
  - Default:    last 1 hour if no time flags are provided

Examples:
  # Search all logs in the last hour (default time range)
  cubeapm logs query '*'

  # Search for errors in a specific service
  cubeapm logs query 'error' --service api-gateway --last 30m

  # Filter by log level
  cubeapm logs query '*' --service payments --level error --last 1h

  # Filter by stream
  cubeapm logs query 'timeout' --stream '{host="web-1"}' --last 2h

  # Use a complex LogsQL expression
  cubeapm logs query 'error AND service:api AND NOT health_check' --last 30m

  # Limit number of results
  cubeapm logs query 'status:500' --limit 50

  # Output as JSON for scripting or piping
  cubeapm logs query 'error' --service api-gateway -o json

  # Output as YAML
  cubeapm logs query '*' --last 10m -o yaml

  # Search with an explicit time range
  cubeapm logs query 'error' --from 2024-01-15T00:00:00Z --to 2024-01-15T12:00:00Z

  # Pipe JSON output to jq for further processing
  cubeapm logs query 'error' -o json | jq '.["_msg"]'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logsql := args[0]

			// Prepend convenience filters to the LogsQL query
			var filters []string
			if service != "" {
				filters = append(filters, "service:"+service)
			}
			if level != "" {
				filters = append(filters, "level:"+level)
			}
			if stream != "" {
				filters = append(filters, "_stream:"+stream)
			}
			if len(filters) > 0 {
				logsql = strings.Join(filters, " AND ") + " AND " + logsql
			}

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			// For table format, collect all results then print
			if cmdutil.OutputFormat == output.FormatTable {
				entries, err := cmdutil.APIClient.QueryLogs(logsql, start, end, limit)
				if err != nil {
					return err
				}

				if len(entries) == 0 {
					fmt.Println("No logs found.")
					return nil
				}

				table := output.TableDef{
					Headers: []string{"TIMESTAMP", "STREAM", "MESSAGE"},
				}
				for _, e := range entries {
					msg := e.Message
					if len(msg) > 120 {
						msg = msg[:117] + "..."
					}
					stream := e.Stream
					if len(stream) > 60 {
						stream = stream[:57] + "..."
					}
					table.Rows = append(table.Rows, []string{
						e.Time,
						stream,
						msg,
					})
				}

				return output.PrintTable(cmdutil.Resolved.NoColor, table)
			}

			// For JSON/YAML, stream output
			formatter := output.NewFormatter(cmdutil.OutputFormat, cmdutil.Resolved.NoColor)
			return cmdutil.APIClient.QueryLogsStream(logsql, start, end, limit, func(entry types.LogEntry) error {
				data := make(map[string]interface{})
				data["_time"] = entry.Time
				data["_stream"] = entry.Stream
				data["_msg"] = entry.Message
				for k, v := range entry.Fields {
					if !strings.HasPrefix(k, "_") {
						data[k] = v
					}
				}
				return formatter.Format(os.Stdout, data)
			})
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of log entries to return")
	cmd.Flags().StringVar(&service, "service", "", "Filter by service name (prepends 'service:<value>' to the query)")
	cmd.Flags().StringVar(&level, "level", "", "Filter by log level: error, warn, info, debug (prepends 'level:<value>' to the query)")
	cmd.Flags().StringVar(&stream, "stream", "", "Filter by log stream (prepends '_stream:<value>' to the query, e.g. '{host=\"web-1\"}')")
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
