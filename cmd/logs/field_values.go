package logs

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newFieldValuesCmd() *cobra.Command {
	var (
		query string
		from  string
		to    string
		last  string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "field-values <field>",
		Short: "List values for a log field",
		Long: `List values for a specific log field and their hit counts.

Returns all unique values for the given field name, along with the count
of log entries containing each value. This is useful for exploring the
cardinality of a field or understanding the distribution of values.

The <field> argument is the field name to query. Use 'cubeapm logs field-names'
to discover available field names.

Optionally filter by a LogsQL query to see values only within matching entries.
Use --limit to cap the number of values returned (default: 100).

Time ranges can be specified as:
  - Relative:   --last 1h  (also: 30m, 2d)
  - RFC3339:    --from 2024-01-15T00:00:00Z
  - Default:    last 1 hour if no time flags are provided

Examples:
  # List all values for the "status" field
  cubeapm logs field-values status --last 1h

  # List hosts in error logs
  cubeapm logs field-values host --query 'error' --limit 50

  # List log levels
  cubeapm logs field-values level --last 24h

  # List services
  cubeapm logs field-values service --last 1h

  # Output as JSON
  cubeapm logs field-values status --last 1h -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			field := args[0]

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			values, err := cmdutil.APIClient.GetLogFieldValues(field, query, start, end, limit)
			if err != nil {
				return err
			}

			if len(values) == 0 {
				fmt.Printf("No values found for field %q.\n", field)
				return nil
			}

			table := output.TableDef{
				Headers: []string{"VALUE", "HITS"},
			}
			for _, v := range values {
				table.Rows = append(table.Rows, []string{
					v.Value,
					strconv.FormatInt(v.Hits, 10),
				})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "LogsQL query to filter values")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of values to return")
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
