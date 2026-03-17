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

Examples:
  cubeapm logs field-values status --last 1h
  cubeapm logs field-values host --query 'error' --limit 50`,
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
