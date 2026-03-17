package traces

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
)

func newServicesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "services",
		Short:   "List all services",
		Long:    "List all services that have reported traces.",
		Aliases: []string{"svc"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := cmdutil.APIClient.GetServices()
			if err != nil {
				return err
			}

			if len(services) == 0 {
				fmt.Println("No services found.")
				return nil
			}

			sort.Strings(services)

			table := output.TableDef{
				Headers: []string{"SERVICE"},
			}
			for _, s := range services {
				table.Rows = append(table.Rows, []string{s})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}
}
