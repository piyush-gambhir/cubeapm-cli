package logs

import (
	"github.com/spf13/cobra"

	logsdelete "github.com/piyush-gambhir/cubeapm-cli/cmd/logs/delete"
)

// NewLogsCmd creates the "logs" parent command.
func NewLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Query and manage logs (VictoriaLogs-compatible)",
		Long: `Query logs using LogsQL syntax, explore fields and streams, and manage log deletions.

CubeAPM stores logs in a VictoriaLogs-compatible format and supports the
LogsQL query language. LogsQL provides keyword search, field filtering,
stream filtering, regex matching, and statistical aggregations.

Subcommands:
  query         Query logs using LogsQL syntax
  hits          Show log volume over time (histogram of matches)
  stats         Execute a LogsQL stats/aggregation query
  streams       List log streams and their entry counts
  field-names   Discover available log field names
  field-values  List values for a specific log field
  delete        Manage log deletion tasks (run, stop, list)

Examples:
  cubeapm logs query 'error AND service:api' --last 30m
  cubeapm logs hits --query 'error' --last 1h --step 5m
  cubeapm logs stats 'error | stats count() by (service)' --last 24h
  cubeapm logs streams --last 1h
  cubeapm logs field-names --last 1h
  cubeapm logs field-values status --last 1h`,
		Aliases: []string{"log"},
	}

	cmd.AddCommand(newQueryCmd())
	cmd.AddCommand(newHitsCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newStreamsCmd())
	cmd.AddCommand(newFieldNamesCmd())
	cmd.AddCommand(newFieldValuesCmd())
	cmd.AddCommand(logsdelete.NewDeleteCmd())

	return cmd
}
