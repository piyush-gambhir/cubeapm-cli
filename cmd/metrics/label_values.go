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

Examples:
  cubeapm metrics label-values job
  cubeapm metrics label-values instance --last 24h`,
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
