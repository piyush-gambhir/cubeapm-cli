package traces

import (
	"github.com/spf13/cobra"
)

// NewTracesCmd creates the "traces" parent command.
func NewTracesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traces",
		Short: "Query and inspect distributed traces",
		Long:  "Search, retrieve, and analyze distributed traces via the Jaeger-compatible API.",
		Aliases: []string{"trace"},
	}

	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newServicesCmd())
	cmd.AddCommand(newOperationsCmd())
	cmd.AddCommand(newDependenciesCmd())

	return cmd
}
