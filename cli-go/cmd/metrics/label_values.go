package metrics

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/timeflag"
)

func newLabelValuesCmd() *cobra.Command {
	var (
		from  string
		to    string
		last  string
		like  string
		match []string
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

  # Narrow to values containing a substring (case-insensitive)
  cubeapm metrics label-values service --like media

  # Scope values to a series selector, e.g. only the PROD environment
  cubeapm metrics label-values service.name --match '{env="PROD"}'

  # Output as JSON
  cubeapm metrics label-values job -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			values, err := cmdutil.APIClient.GetLabelValues(label, match, start, end)
			if err != nil {
				return err
			}

			// --like does a case-insensitive substring match. Service
			// inventories often have case-inconsistent spellings
			// (MEDIA-SERVICE, Media-Service, media_service), narrowing to
			// a family is more useful than eyeballing a long sorted list.
			if like != "" {
				needle := strings.ToLower(like)
				filtered := values[:0]
				for _, v := range values {
					if strings.Contains(strings.ToLower(v), needle) {
						filtered = append(filtered, v)
					}
				}
				values = filtered
			}

			if len(values) == 0 && cmdutil.OutputFormat == output.FormatTable {
				if like != "" {
					fmt.Printf("No values matching %q found for label %q.\n", like, label)
				} else {
					fmt.Printf("No values found for label %q.\n", label)
				}
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

	cmd.Flags().StringVar(&like, "like", "", "Case-insensitive substring filter applied to returned values")
	cmd.Flags().StringArrayVar(&match, "match", nil, `Series selector to scope returned values; repeatable and ORed (e.g. --match '{env="PROD"}')`)
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
