package metrics

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newSeriesCmd() *cobra.Command {
	var (
		from  string
		to    string
		last  string
		match []string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "series",
		Short: "Find time series matching a label set",
		Long: `Find time series matching one or more series selectors.

Queries the Prometheus-compatible /api/v1/series endpoint to discover
time series that match the given label matchers. At least one --match
selector is required.

Selectors use PromQL label matching syntax:
  metric_name                     - match by metric name
  metric_name{label="value"}     - exact label match
  metric_name{label=~"regex"}    - regex label match
  metric_name{label!="value"}    - negative match
  {job="api"}                    - match any metric with label

The --limit flag restricts the number of series returned (default: unlimited).
Use this to avoid overwhelming output on high-cardinality metrics.

Time ranges can be specified as:
  - Relative:   --last 1h  (also: 30m, 2d)
  - RFC3339:    --from 2024-01-15T10:00:00Z
  - Default:    last 1 hour if no time flags are provided

Examples:
  # Find series matching a metric name
  cubeapm metrics series --match 'up{job="api"}'

  # Find series matching multiple selectors
  cubeapm metrics series --match 'http_requests_total' --match 'process_cpu_seconds_total'

  # Limit the number of returned series
  cubeapm metrics series --match 'http_requests_total' --limit 50

  # Find series with regex matching
  cubeapm metrics series --match '{__name__=~"http_.*"}' --last 24h

  # Output as JSON
  cubeapm metrics series --match 'up' -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(match) == 0 {
				return fmt.Errorf("at least one --match selector is required")
			}

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			series, err := cmdutil.APIClient.GetSeries(match, start, end, limit)
			if err != nil {
				return err
			}

			if len(series) == 0 {
				fmt.Println("No matching series found.")
				return nil
			}

			table := output.TableDef{
				Headers: []string{"SERIES"},
			}
			for _, s := range series {
				table.Rows = append(table.Rows, []string{formatLabelSet(s)})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	timeflag.AddTimeFlags(cmd, &from, &to, &last)
	cmd.Flags().StringArrayVar(&match, "match", nil, "Series selector (can be specified multiple times)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of series to return (0 = unlimited)")

	return cmd
}

func formatLabelSet(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var name string
	parts := make([]string, 0, len(labels))
	for _, k := range keys {
		if k == "__name__" {
			name = labels[k]
			continue
		}
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, labels[k]))
	}

	result := name
	if len(parts) > 0 {
		result += "{" + strings.Join(parts, ", ") + "}"
	}
	return result
}
