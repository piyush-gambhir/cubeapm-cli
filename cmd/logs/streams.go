package logs

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newStreamsCmd() *cobra.Command {
	var (
		query string
		from  string
		to    string
		last  string
	)

	cmd := &cobra.Command{
		Use:   "streams",
		Short: "List log streams",
		Long: `List log streams and their entry counts.

Examples:
  cubeapm logs streams --last 1h
  cubeapm logs streams --query 'error' --last 24h`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			streams, err := cmdutil.APIClient.GetLogStreams(query, start, end)
			if err != nil {
				return err
			}

			if len(streams) == 0 {
				fmt.Println("No streams found.")
				return nil
			}

			table := output.TableDef{
				Headers: []string{"STREAM", "ENTRIES"},
			}
			for _, s := range streams {
				table.Rows = append(table.Rows, []string{
					s.Stream,
					strconv.FormatInt(s.Entries, 10),
				})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "LogsQL query to filter streams")
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
