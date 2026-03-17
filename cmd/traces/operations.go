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
		Use:     "operations <service>",
		Short:   "List operations for a service",
		Long:    "List all operations (endpoints/methods) reported by a service.",
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
