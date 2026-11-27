package logs

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/timeflag"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/types"
)

func prependFilters(filters []string, logsql string) string {
	if len(filters) == 0 {
		return logsql
	}
	prefix := strings.Join(filters, " AND ") + " AND "
	filter, pipes, comment := splitPipes(logsql)
	filter = strings.TrimSpace(filter)
	switch {
	case filter == "*":
	case comment:
		filter = "(" + filter + "\n)"
	default:
		filter = "(" + filter + ")"
	}
	if pipes == "" {
		return prefix + filter
	}
	return prefix + filter + " " + pipes
}

func splitPipes(logsql string) (string, string, bool) {
	var quote rune
	escaped, inComment, comment := false, false, false
	depth := 0
	for i, r := range logsql {
		switch {
		case inComment:
			if r == '\n' {
				inComment = false
			}
		case escaped:
			escaped = false
		case quote != 0:
			if r == '\\' && quote != '`' {
				escaped = true
			} else if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'' || r == '`':
			quote = r
		case r == '#':
			inComment, comment = true, true
		case r == '(' || r == '[':
			depth++
		case (r == ')' || r == ']') && depth > 0:
			depth--
		case r == '|' && depth == 0:
			return logsql[:i], logsql[i:], comment
		}
	}
	return logsql, "", comment
}

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
results are collected and displayed as a table. In JSON/YAML mode, entries
are streamed one object at a time instead of one array: -o json emits a
sequence of JSON objects, and -o yaml emits one YAML document per entry,
separated by "---".

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
			logsql = prependFilters(filters, logsql)

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			if cmdutil.OutputFormat == output.FormatTable {
				entries, err := cmdutil.APIClient.QueryLogs(logsql, start, end, limit)
				if err != nil {
					return err
				}

				if len(entries) == 0 {
					fmt.Println("No logs found.")
					if service != "" {
						fmt.Fprintf(os.Stderr, "\nHint: the --service filter narrowed by `service:%s`. If you expected results, check spelling with:\n  cubeapm logs field-values service --last 24h\n  cubeapm logs field-values service.name --last 24h\nSome services don't emit logs to CubeAPM at all, verify there.\n", service)
					}
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

			out := cmd.OutOrStdout()
			formatter := output.NewFormatter(cmdutil.OutputFormat, cmdutil.Resolved.NoColor)
			first := true
			return cmdutil.APIClient.QueryLogsStream(logsql, start, end, limit, func(entry types.LogEntry) error {
				if cmdutil.OutputFormat == output.FormatYAML && !first {
					if _, err := io.WriteString(out, "---\n"); err != nil {
						return err
					}
				}
				first = false
				data := make(map[string]interface{})
				data["_time"] = entry.Time
				data["_stream"] = entry.Stream
				data["_msg"] = entry.Message
				for k, v := range entry.Fields {
					if !strings.HasPrefix(k, "_") {
						data[k] = v
					}
				}
				return formatter.Format(out, data)
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
