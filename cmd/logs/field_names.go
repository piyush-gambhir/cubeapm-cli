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

Examples:
  cubeapm logs field-names --last 1h
  cubeapm logs field-names --query 'service:api' --last 24h`,
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
