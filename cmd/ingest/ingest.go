package ingest

import (
	"github.com/spf13/cobra"
)

// NewIngestCmd creates the "ingest" parent command.
func NewIngestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Push data to CubeAPM ingest endpoints",
		Long:  "Ingest metrics and logs data in various formats via the ingest port.",
	}

	cmd.AddCommand(newIngestMetricsCmd())
	cmd.AddCommand(newIngestLogsCmd())

	return cmd
}
