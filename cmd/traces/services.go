package traces

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newServicesCmd() *cobra.Command {
	var (
		env  string
		from string
		to   string
		last string
	)

	cmd := &cobra.Command{
		Use:   "services",
		Short: "List all services",
		Long: `List all services that have reported telemetry to CubeAPM.

Returns a sorted list of service names. This is typically the first command
you run to discover what services are available before searching for traces
or listing operations.

Service names correspond to the service.name resource attribute set by the
OpenTelemetry SDK or tracing agent in your application.

With --env, list only services seen in a specific environment (e.g. PROD,
UAT). The environment list is derived from the metrics layer (label-values
filtered by the env / cube.environment labels), so a service that emits only
traces and no metrics in that environment may not appear. Without --env, all
known services are listed.

Examples:
  # List all services
  cubeapm traces services

  # List only PROD services (env labels are upper-case: PROD, UAT, ...)
  cubeapm traces services --env PROD

  # Scope --env to a recent window (defaults to the last hour)
  cubeapm traces services --env PROD --last 24h

  # Output as JSON
  cubeapm traces services -o json

  # Use the short alias
  cubeapm traces svc`,
		Aliases: []string{"svc"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				services []string
				err      error
			)
			if env != "" {
				start, end, terr := timeflag.ResolveTimeRange(from, to, last)
				if terr != nil {
					return terr
				}
				services, err = cmdutil.APIClient.GetServicesByEnv(env, start, end)
			} else {
				services, err = cmdutil.APIClient.GetServices()
			}
			if err != nil {
				return err
			}

			if len(services) == 0 && cmdutil.OutputFormat == output.FormatTable {
				if env != "" {
					fmt.Printf("No services found in environment %q.\n", env)
				} else {
					fmt.Println("No services found.")
				}
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

	cmd.Flags().StringVar(&env, "env", "", "List only services seen in this environment, e.g. PROD or UAT (metrics-derived)")
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}
