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
		Use:   "services",
		Short: "List all services",
		Long: `List all services that have reported traces to CubeAPM.

Returns a sorted list of service names. This is typically the first command
you run to discover what services are available before searching for traces
or listing operations.

Service names correspond to the service.name resource attribute set by
the OpenTelemetry SDK or tracing library in your application.

Examples:
  # List all services
  cubeapm traces services

  # List services and output as JSON
  cubeapm traces services -o json

  # Use the short alias
  cubeapm traces svc`,
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
