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
	)

	cmd := &cobra.Command{
		Use:   "series",
		Short: "Find time series matching a label set",
		Long: `Find time series matching one or more series selectors.

Examples:
  cubeapm metrics series --match 'up{job="api"}'
  cubeapm metrics series --match 'http_requests_total' --match 'process_cpu_seconds_total'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(match) == 0 {
				return fmt.Errorf("at least one --match selector is required")
			}

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			series, err := cmdutil.APIClient.GetSeries(match, start, end)
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
