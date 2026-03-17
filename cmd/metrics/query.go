package metrics

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/client"
	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

func newQueryCmd() *cobra.Command {
	var queryTime string

	cmd := &cobra.Command{
		Use:   "query <promql>",
		Short: "Execute an instant PromQL query",
		Long: `Execute an instant PromQL query and display the results.

Evaluates a PromQL expression at a single point in time (default: now).
Returns the current value of each matching time series.

The <promql> argument is a PromQL expression. Common PromQL patterns:
  - Simple metric:       'up'
  - Rate of counter:     'rate(http_requests_total[5m])'
  - Aggregation:         'sum by (service) (rate(requests_total[5m]))'
  - Label filtering:     'http_requests_total{method="GET", status="200"}'
  - Math operations:     'rate(errors_total[5m]) / rate(requests_total[5m])'
  - Histogram quantile:  'histogram_quantile(0.99, rate(http_duration_seconds_bucket[5m]))'

The --time flag specifies the evaluation timestamp. If omitted, "now" is used.
Supported time formats: RFC3339, Unix timestamp, or relative (now-1h, -30m).

In table mode (default), displays metric labels, value, and timestamp.
In JSON/YAML mode, returns the raw Prometheus API response.

Examples:
  # Check which targets are up
  cubeapm metrics query 'up'

  # Compute request rate per service
  cubeapm metrics query 'sum by (service) (rate(http_requests_total[5m]))'

  # Query at a specific time in the past
  cubeapm metrics query 'up' --time now-1h

  # Query at an exact RFC3339 timestamp
  cubeapm metrics query 'up' --time 2024-01-15T10:00:00Z

  # Get error rate as a percentage
  cubeapm metrics query 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100'

  # Output as JSON for scripting
  cubeapm metrics query 'up' -o json

  # Compute p99 latency
  cubeapm metrics query 'histogram_quantile(0.99, sum by (le) (rate(http_duration_seconds_bucket[5m])))'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			promql := args[0]

			var t time.Time
			if queryTime != "" {
				var err error
				t, err = timeflag.ParseTime(queryTime)
				if err != nil {
					return fmt.Errorf("parsing --time: %w", err)
				}
			}

			result, err := cmdutil.APIClient.QueryInstant(promql, t)
			if err != nil {
				return err
			}

			// For non-table output, return raw result
			if cmdutil.OutputFormat != output.FormatTable {
				return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, result)
			}

			return renderInstantResult(result, cmdutil.Resolved.NoColor)
		},
	}

	cmd.Flags().StringVar(&queryTime, "time", "", "Evaluation time (RFC3339, Unix, or relative)")

	return cmd
}

func renderInstantResult(result *types.MetricsResult, noColor bool) error {
	switch result.Data.ResultType {
	case "vector":
		var samples types.VectorResult
		if err := json.Unmarshal(result.Data.Result, &samples); err != nil {
			return fmt.Errorf("parsing vector result: %w", err)
		}

		table := output.TableDef{
			Headers: []string{"METRIC", "VALUE", "TIMESTAMP"},
		}
		for _, s := range samples {
			metric := client.FormatMetricLabels(s.Metric)
			ts := time.Unix(int64(s.Value.Timestamp()), 0).Format(time.RFC3339)
			table.Rows = append(table.Rows, []string{
				metric,
				s.Value.Value(),
				ts,
			})
		}
		return output.PrintTable(noColor, table)

	case "scalar":
		var pair types.SamplePair
		if err := json.Unmarshal(result.Data.Result, &pair); err != nil {
			return fmt.Errorf("parsing scalar result: %w", err)
		}
		table := output.TableDef{
			Headers: []string{"VALUE"},
			Rows:    [][]string{{pair.Value()}},
		}
		return output.PrintTable(noColor, table)

	case "string":
		var pair types.SamplePair
		if err := json.Unmarshal(result.Data.Result, &pair); err != nil {
			return fmt.Errorf("parsing string result: %w", err)
		}
		fmt.Println(pair.Value())
		return nil

	default:
		return output.Print(output.FormatJSON, noColor, result)
	}
}
