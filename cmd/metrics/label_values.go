package metrics

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newLabelValuesCmd() *cobra.Command {
	var (
		from string
		to   string
		last string
	)

	cmd := &cobra.Command{
		Use:   "label-values <label>",
		Short: "List values for a metric label",
		Long: `List all values for a specific metric label.

Queries the Prometheus-compatible /api/v1/label/<label>/values endpoint
to return a sorted list of all values seen for the given label name.

The <label> argument is the label name to query. Use 'cubeapm metrics labels'
to discover available label names.

Common uses:
  - List all job names:       cubeapm metrics label-values job
  - List all instances:       cubeapm metrics label-values instance
  - List all metric names:    cubeapm metrics label-values __name__
  - List all services:        cubeapm metrics label-values service

Time ranges can be specified to limit the scope:
  - Relative:   --last 24h
  - RFC3339:    --from 2024-01-15T00:00:00Z
  - Default:    last 1 hour if no time flags are provided

Examples:
  # List all values for the "job" label
  cubeapm metrics label-values job

  # List all instances seen in the last 24 hours
  cubeapm metrics label-values instance --last 24h

  # List all metric names
  cubeapm metrics label-values __name__

  # Output as JSON
  cubeapm metrics label-values job -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			values, err := cmdutil.APIClient.GetLabelValues(label, start, end)
			if err != nil {
				return err
			}

			if len(values) == 0 {
				fmt.Printf("No values found for label %q.\n", label)
				return nil
			}

			sort.Strings(values)

			table := output.TableDef{
				Headers: []string{"VALUE"},
			}
			for _, v := range values {
				table.Rows = append(table.Rows, []string{v})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
