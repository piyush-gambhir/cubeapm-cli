package metrics

import (
	"github.com/spf13/cobra"
)

// NewMetricsCmd creates the "metrics" parent command.
func NewMetricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Query Prometheus-compatible metrics",
		Long:  "Execute PromQL queries and explore metric labels and series.",
		Aliases: []string{"metric"},
	}

	cmd.AddCommand(newQueryCmd())
	cmd.AddCommand(newQueryRangeCmd())
	cmd.AddCommand(newLabelsCmd())
	cmd.AddCommand(newLabelValuesCmd())
	cmd.AddCommand(newSeriesCmd())

	return cmd
}
