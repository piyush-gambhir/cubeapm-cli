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
		Long:  "Query logs using LogsQL syntax, explore fields and streams, and manage log deletions.",
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
