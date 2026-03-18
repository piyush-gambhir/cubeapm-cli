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
		Use:         "logs",
		Short:       "Push log data to CubeAPM",
		Annotations: map[string]string{"mutates": "true"},
		Long: `Push log data in supported formats to the CubeAPM ingest endpoint.

Reads log data from a file or stdin and sends it to the CubeAPM ingest
port (default: 3130). The data must be in one of the supported formats.

Supported formats:
  jsonline  - Newline-delimited JSON (default). Each line is a JSON object:
              {"_time":"2024-01-15T10:00:00Z","_msg":"request completed","service":"api"}
  otlp      - OpenTelemetry Protocol (protobuf binary)
  loki      - Loki push API format (JSON with streams/values arrays)
  elastic   - Elasticsearch bulk format (NDJSON with action/document pairs)

Data source:
  --file <path>  Read from a file
  --file -       Read from stdin (default)
  (no --file)    Read from stdin (pipe or redirect required)

Examples:
  # Ingest JSON line logs from a file
  cubeapm ingest logs --format jsonline --file logs.jsonl

  # Pipe JSON line logs from stdin
  cat logs.jsonl | cubeapm ingest logs --format jsonline

  # Ingest from a log file using a converter
  cat app.log | my-log-to-json | cubeapm ingest logs --format jsonline

  # Ingest OTLP protobuf data
  cubeapm ingest logs --format otlp --file logs.pb

  # Ingest Loki-format data
  cubeapm ingest logs --format loki --file loki-push.json

  # Ingest Elasticsearch bulk format
  cubeapm ingest logs --format elastic --file elastic-bulk.ndjson

  # Use explicit stdin marker
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

			if !cmdutil.Quiet {
				fmt.Println("Logs ingested successfully.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "jsonline", "Data format: jsonline, otlp, loki, elastic")
	cmd.Flags().StringVar(&file, "file", "-", "File path (use - for stdin)")

	return cmd
}
