package metrics

import (
	"github.com/spf13/cobra"
)

// NewMetricsCmd creates the "metrics" parent command.
func NewMetricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Query Prometheus-compatible metrics",
		Long: `Execute PromQL queries and explore metric labels and series.

CubeAPM exposes a Prometheus-compatible metrics API. You can use standard
PromQL syntax for instant queries, range queries, and series exploration.

Subcommands:
  query         Execute an instant PromQL query at a single point in time
  query-range   Execute a range PromQL query over a time window
  labels        List all available metric label names
  label-values  List values for a specific metric label
  series        Find time series matching label selectors

Examples:
  cubeapm metrics query 'up'
  cubeapm metrics query-range 'rate(http_requests_total[5m])' --last 1h
  cubeapm metrics labels --last 24h
  cubeapm metrics label-values job
  cubeapm metrics series --match 'http_requests_total'`,
		Aliases: []string{"metric"},
	}

	cmd.AddCommand(newQueryCmd())
	cmd.AddCommand(newQueryRangeCmd())
	cmd.AddCommand(newLabelsCmd())
	cmd.AddCommand(newLabelValuesCmd())
	cmd.AddCommand(newSeriesCmd())

	return cmd
}
