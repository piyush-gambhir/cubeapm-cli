package ingest

import (
	"github.com/spf13/cobra"
)

// NewIngestCmd creates the "ingest" parent command.
func NewIngestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Push data to CubeAPM ingest endpoints",
		Long: `Push metrics and logs data to CubeAPM via the ingest port.

The ingest commands send data to CubeAPM's ingest endpoint (default port: 3130).
Data can be read from a file or piped via stdin.

Subcommands:
  metrics   Push metrics data (Prometheus exposition format or OTLP)
  logs      Push log data (JSON lines, OTLP, Loki, or Elasticsearch format)

Examples:
  cubeapm ingest metrics --format prometheus --file metrics.txt
  cat logs.jsonl | cubeapm ingest logs --format jsonline
  cubeapm ingest logs --format otlp --file logs.pb`,
	}

	cmd.AddCommand(newIngestMetricsCmd())
	cmd.AddCommand(newIngestLogsCmd())

	return cmd
}
