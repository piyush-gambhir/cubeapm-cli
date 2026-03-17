package ingest

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
)

func newIngestLogsCmd() *cobra.Command {
	var (
		format string
		file   string
	)

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Push log data to CubeAPM",
		Long: `Push log data in supported formats to the CubeAPM ingest endpoint.

Supported formats:
  jsonline  - Newline-delimited JSON
  otlp      - OpenTelemetry Protocol (protobuf)
  loki      - Loki push format
  elastic   - Elasticsearch bulk format

Examples:
  cubeapm ingest logs --format jsonline --file logs.jsonl
  cat logs.jsonl | cubeapm ingest logs --format jsonline --file -`,
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

			if err := cmdutil.APIClient.IngestLogs(format, reader); err != nil {
				return err
			}

			fmt.Println("Logs ingested successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "jsonline", "Data format: jsonline, otlp, loki, elastic")
	cmd.Flags().StringVar(&file, "file", "-", "File path (use - for stdin)")

	return cmd
}
