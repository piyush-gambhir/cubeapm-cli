package traces

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
)

func newOperationsCmd() *cobra.Command {
	var spanKind string

	cmd := &cobra.Command{
		Use:   "operations <service>",
		Short: "List operations for a service",
		Long: `List all operations (endpoints/methods) reported by a service.

Queries the Jaeger-compatible API to retrieve all known operations for the
given service. Each operation is listed with its span kind (client, server,
producer, consumer, or internal).

The <service> argument is required and must be the exact service name as
reported in traces. Use 'cubeapm traces services' to discover available
service names.

Optionally filter by span kind to see only server-side or client-side operations.

Examples:
  # List all operations for a service
  cubeapm traces operations api-gateway

  # List only server-side operations
  cubeapm traces operations api-gateway --span-kind server

  # List client-side operations (outgoing calls)
  cubeapm traces operations payments --span-kind client

  # Output as JSON
  cubeapm traces operations api-gateway -o json

  # Use the short alias
  cubeapm traces ops api-gateway`,
		Aliases: []string{"ops"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := args[0]

			operations, err := cmdutil.APIClient.GetOperations(service, spanKind)
			if err != nil {
				return err
			}

			if len(operations) == 0 {
				fmt.Printf("No operations found for service %q.\n", service)
				return nil
			}

			table := output.TableDef{
				Headers: []string{"OPERATION", "SPAN_KIND"},
			}
			for _, op := range operations {
				table.Rows = append(table.Rows, []string{op.Name, op.SpanKind})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	cmd.Flags().StringVar(&spanKind, "span-kind", "", "Filter by span kind")

	return cmd
}
