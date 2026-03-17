package metrics

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newLabelsCmd() *cobra.Command {
	var (
		from string
		to   string
		last string
	)

	cmd := &cobra.Command{
		Use:   "labels",
		Short: "List metric label names",
		Long: `List all available metric label names.

Queries the Prometheus-compatible /api/v1/labels endpoint to return a sorted
list of all label names present in the stored metrics. This is useful for
discovering what labels are available before constructing PromQL queries.

Common standard labels include:
  __name__   - metric name
  job        - scrape job name
  instance   - scrape target instance
  service    - service name (if set by instrumentation)

Time ranges can be specified to limit the scope of the label search:
  - Relative:   --last 24h
  - RFC3339:    --from 2024-01-15T00:00:00Z
  - Default:    last 1 hour if no time flags are provided

Examples:
  # List all label names
  cubeapm metrics labels

  # List labels seen in the last 24 hours
  cubeapm metrics labels --last 24h

  # Output as JSON
  cubeapm metrics labels -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			labels, err := cmdutil.APIClient.GetLabels(start, end)
			if err != nil {
				return err
			}

			if len(labels) == 0 {
				fmt.Println("No labels found.")
				return nil
			}

			sort.Strings(labels)

			table := output.TableDef{
				Headers: []string{"LABEL"},
			}
			for _, l := range labels {
				table.Rows = append(table.Rows, []string{l})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
