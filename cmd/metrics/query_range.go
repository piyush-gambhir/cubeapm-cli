package metrics

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/client"
	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

func newQueryRangeCmd() *cobra.Command {
	var (
		from string
		to   string
		last string
		step string
	)

	cmd := &cobra.Command{
		Use:   "query-range <promql>",
		Short: "Execute a range PromQL query",
		Long: `Execute a range PromQL query and display the results as a time series.

Examples:
  cubeapm metrics query-range 'rate(http_requests_total[5m])' --last 1h --step 1m
  cubeapm metrics query-range 'sum by (service) (up)' --from -2h --to -1h`,
		Aliases: []string{"range"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			promql := args[0]

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			result, err := cmdutil.APIClient.QueryRange(promql, start, end, step)
			if err != nil {
				return err
			}

			// For non-table output, return raw result
			if cmdutil.OutputFormat != output.FormatTable {
				return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, result)
			}

			return renderRangeResult(result, cmdutil.Resolved.NoColor)
		},
	}

	timeflag.AddTimeFlags(cmd, &from, &to, &last)
	cmd.Flags().StringVar(&step, "step", "", "Query resolution step (e.g., 15s, 1m, 5m)")

	return cmd
}

func renderRangeResult(result *types.MetricsResult, noColor bool) error {
	if result.Data.ResultType != "matrix" {
		return output.Print(output.FormatJSON, noColor, result)
	}

	var series types.MatrixResult
	if err := json.Unmarshal(result.Data.Result, &series); err != nil {
		return fmt.Errorf("parsing matrix result: %w", err)
	}

	if len(series) == 0 {
		fmt.Println("No data points returned.")
		return nil
	}

	table := output.TableDef{
		Headers: []string{"METRIC", "SAMPLES", "VALUES"},
	}

	for _, s := range series {
		metric := client.FormatMetricLabels(s.Metric)
		sampleCount := fmt.Sprintf("%d", len(s.Values))

		// Show first few and last few values
		var valParts []string
		maxShow := 5
		if len(s.Values) <= maxShow*2 {
			for _, v := range s.Values {
				ts := time.Unix(int64(v.Timestamp()), 0).Format("15:04:05")
				valParts = append(valParts, fmt.Sprintf("%s=%s", ts, v.Value()))
			}
		} else {
			for i := 0; i < maxShow; i++ {
				v := s.Values[i]
				ts := time.Unix(int64(v.Timestamp()), 0).Format("15:04:05")
				valParts = append(valParts, fmt.Sprintf("%s=%s", ts, v.Value()))
			}
			valParts = append(valParts, "...")
			for i := len(s.Values) - maxShow; i < len(s.Values); i++ {
				v := s.Values[i]
				ts := time.Unix(int64(v.Timestamp()), 0).Format("15:04:05")
				valParts = append(valParts, fmt.Sprintf("%s=%s", ts, v.Value()))
			}
		}

		table.Rows = append(table.Rows, []string{
			metric,
			sampleCount,
			strings.Join(valParts, " "),
		})
	}

	return output.PrintTable(noColor, table)
}
