package ingest

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
)

func newIngestMetricsCmd() *cobra.Command {
	var (
		format string
		file   string
	)

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Push metrics data to CubeAPM",
		Long: `Push metrics data in supported formats to the CubeAPM ingest endpoint.

Reads metrics data from a file or stdin and sends it to the CubeAPM ingest
port (default: 3130). The data must be in one of the supported formats.

Supported formats:
  prometheus  - Prometheus exposition/text format (default)
                Lines like: http_requests_total{method="GET"} 1234 1705312800000
  otlp        - OpenTelemetry Protocol (protobuf binary)

Data source:
  --file <path>  Read from a file
  --file -       Read from stdin (default)
  (no --file)    Read from stdin (pipe or redirect required)

The ingest endpoint URL is: http://<server>:<ingest-port>/api/v1/import/prometheus
for Prometheus format, or http://<server>:<ingest-port>/v1/metrics for OTLP.

Examples:
  # Ingest metrics from a file (Prometheus format)
  cubeapm ingest metrics --format prometheus --file metrics.txt

  # Pipe metrics from stdin
  cat metrics.txt | cubeapm ingest metrics --format prometheus

  # Pipe from a curl command
  curl -s http://localhost:9090/metrics | cubeapm ingest metrics --format prometheus

  # Ingest OTLP protobuf data
  cubeapm ingest metrics --format otlp --file metrics.pb

  # Use explicit stdin marker
  cat metrics.txt | cubeapm ingest metrics --format prometheus --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var reader io.Reader

			if file == "-" || file == "" {
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) != 0 {
					return fmt.Errorf("no data provided: pipe data via stdin or use --file")
				}
				reader = os.Stdin
			} else {
				f, err := os.Open(file)
				if err != nil {
					return fmt.Errorf("opening file: %w", err)
				}
				defer f.Close()
				reader = f
			}

			if err := cmdutil.APIClient.IngestMetrics(format, reader); err != nil {
				return err
			}

			fmt.Println("Metrics ingested successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "prometheus", "Data format: prometheus, otlp")
	cmd.Flags().StringVar(&file, "file", "-", "File path (use - for stdin)")

	return cmd
}
