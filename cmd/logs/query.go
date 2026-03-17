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
		from  string
		to    string
		last  string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "query <logsql>",
		Short: "Query logs using LogsQL",
		Long: `Query logs using LogsQL syntax. Results are streamed by default.

Examples:
  cubeapm logs query '*'
  cubeapm logs query 'error AND service:api' --last 30m
  cubeapm logs query '_stream:{host="web-1"}' --limit 100
  cubeapm logs query 'status:500' -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logsql := args[0]

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
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
