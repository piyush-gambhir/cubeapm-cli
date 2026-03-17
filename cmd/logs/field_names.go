package logs

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newFieldNamesCmd() *cobra.Command {
	var (
		query string
		from  string
		to    string
		last  string
	)

	cmd := &cobra.Command{
		Use:   "field-names",
		Short: "Discover available log fields",
		Long: `List all log field names and their hit counts.

Returns the names of all fields present in matching log entries, along with
the count of entries containing each field. This is useful for schema
discovery - understanding what structured data is available in your logs
before constructing queries.

Standard fields include:
  _time     - timestamp
  _msg      - log message body
  _stream   - stream identifier (labels like host, container)

Additional fields depend on your log format (e.g., level, service, trace_id,
http.method, http.status_code, etc.).

Optionally filter by a LogsQL query to see fields only within matching entries.

Time ranges can be specified as:
  - Relative:   --last 1h  (also: 30m, 2d)
  - RFC3339:    --from 2024-01-15T00:00:00Z
  - Default:    last 1 hour if no time flags are provided

Examples:
  # List all field names in the last hour
  cubeapm logs field-names --last 1h

  # List fields present in logs for a specific service
  cubeapm logs field-names --query 'service:api' --last 24h

  # List fields in error logs
  cubeapm logs field-names --query 'level:error' --last 1h

  # Use the short alias
  cubeapm logs fields --last 1h

  # Output as JSON
  cubeapm logs field-names --last 1h -o json`,
		Aliases: []string{"fields"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			fields, err := cmdutil.APIClient.GetLogFieldNames(query, start, end)
			if err != nil {
				return err
			}

			if len(fields) == 0 {
				fmt.Println("No fields found.")
				return nil
			}

			table := output.TableDef{
				Headers: []string{"FIELD", "HITS"},
			}
			for _, f := range fields {
				table.Rows = append(table.Rows, []string{
					f.Name,
					strconv.FormatInt(f.Hits, 10),
				})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "LogsQL query to filter fields")
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
