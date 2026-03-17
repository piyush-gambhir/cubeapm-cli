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

Supported formats:
  prometheus  - Prometheus exposition format
  otlp        - OpenTelemetry Protocol (protobuf)

Examples:
  cubeapm ingest metrics --format prometheus --file metrics.txt
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
