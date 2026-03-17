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
		Long:  "List all available metric label names.",
		Args:  cobra.NoArgs,
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
