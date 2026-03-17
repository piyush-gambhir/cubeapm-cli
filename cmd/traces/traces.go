package traces

import (
	"github.com/spf13/cobra"
)

// NewTracesCmd creates the "traces" parent command.
func NewTracesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traces",
		Short: "Query and inspect distributed traces",
		Long: `Search, retrieve, and analyze distributed traces via the Jaeger-compatible API.

CubeAPM stores traces in Jaeger format. Each trace consists of one or more
spans representing units of work across services. Use the subcommands to
search traces, view individual traces, list services and operations, and
explore service dependencies.

Subcommands:
  search        Search for traces matching filters (service, status, duration, tags)
  get           Retrieve a specific trace by its trace ID
  services      List all services that have reported traces
  operations    List operations (endpoints/methods) for a service
  dependencies  Show the service dependency graph

Examples:
  cubeapm traces search --service api-gateway --last 1h
  cubeapm traces get abc123def456
  cubeapm traces services
  cubeapm traces operations api-gateway
  cubeapm traces dependencies --last 24h`,
		Aliases: []string{"trace"},
	}

	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newServicesCmd())
	cmd.AddCommand(newOperationsCmd())
	cmd.AddCommand(newDependenciesCmd())

	return cmd
}
