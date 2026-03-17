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

Examples:
  cubeapm metrics query 'up'
  cubeapm metrics query 'rate(http_requests_total[5m])' --time now-1h
  cubeapm metrics query 'sum by (service) (rate(requests_total[5m]))'`,
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
