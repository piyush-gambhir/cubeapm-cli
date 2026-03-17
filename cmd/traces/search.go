package traces

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newSearchCmd() *cobra.Command {
	var (
		service     string
		env         string
		query       string
		from        string
		to          string
		last        string
		limit       int
		spanKind    string
		status      string
		minDuration string
		maxDuration string
		tags        []string
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for traces",
		Long: `Search for traces matching the given criteria.

Queries the Jaeger-compatible traces API. Results include trace ID, service,
operation, duration, status, and timestamp for each matching trace's root span.

Filters:
  --service       Filter by service name (e.g., "api-gateway", "payments")
  --env           Filter by environment tag (e.g., "production", "staging")
  --query         Filter by operation name (e.g., "GET /api/users")
  --status        Filter by span status: "error" or "ok"
  --min-duration  Filter traces slower than this duration (e.g., "500ms", "1s", "100us")
  --max-duration  Filter traces faster than this duration (e.g., "5s", "10s")
  --tags          Filter by span tag key=value pairs (can be specified multiple times)
  --span-kind     Filter by span kind: client, server, producer, consumer, internal
  --limit         Maximum number of traces to return (default: 20)

Time ranges can be specified as:
  - Relative:   --last 1h  (also: 30m, 2d, 1d12h)
  - RFC3339:    --from 2024-01-15T10:00:00Z
  - Unix:       --from 1705312800
  - Default:    last 1 hour if no time flags are provided

Examples:
  # Search traces for a service in the last hour
  cubeapm traces search --service api-gateway --last 1h

  # Find slow traces (>500ms) with errors
  cubeapm traces search --service payments --min-duration 500ms --status error

  # Search with a maximum duration filter
  cubeapm traces search --service auth --max-duration 100ms --last 2h

  # Filter by operation name
  cubeapm traces search --service api-gateway --query "GET /api/users" --last 1h

  # Filter by span tags
  cubeapm traces search --service api-gateway --tags "http.method=POST" --tags "http.status_code=500"

  # Filter by environment and span kind
  cubeapm traces search --service payments --env production --span-kind server

  # Search with a custom time range
  cubeapm traces search --service auth --from 2024-01-15T00:00:00Z --to 2024-01-15T12:00:00Z

  # Return more results
  cubeapm traces search --service api-gateway --limit 100

  # Output as JSON for scripting
  cubeapm traces search --service api-gateway -o json

  # Pipe to jq for further processing
  cubeapm traces search --service api-gateway -o json | jq '.[].TRACE_ID'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			// Build tags JSON from --tags and --status flags
			tagsMap := make(map[string]string)
			if env != "" {
				tagsMap["environment"] = env
			}
			if status != "" {
				statusLower := strings.ToLower(status)
				if statusLower == "error" {
					tagsMap["error"] = "true"
				} else if statusLower == "ok" {
					tagsMap["otel.status_code"] = "OK"
				}
			}
			for _, t := range tags {
				parts := strings.SplitN(t, "=", 2)
				if len(parts) == 2 {
					tagsMap[parts[0]] = parts[1]
				}
			}

			results, err := cmdutil.APIClient.SearchTraces(service, tagsMap, query, start, end, limit, spanKind, minDuration, maxDuration)
			if err != nil {
				return err
			}

			if len(results) == 0 {
				fmt.Println("No traces found.")
				return nil
			}

			table := output.TableDef{
				Headers: []string{"TRACE_ID", "SERVICE", "OPERATION", "DURATION", "STATUS", "TIMESTAMP"},
			}

			for _, trace := range results {
				if len(trace.Spans) == 0 {
					continue
				}
				// Find the root span (no parent reference or first span)
				rootSpan := trace.Spans[0]
				for _, span := range trace.Spans {
					if len(span.References) == 0 {
						rootSpan = span
						break
					}
				}

				// Get service name from processes
				serviceName := ""
				if proc, ok := trace.Processes[rootSpan.ProcessID]; ok {
					serviceName = proc.ServiceName
				}

				// Get status from tags
				spanStatus := "OK"
				for _, tag := range rootSpan.Tags {
					if tag.Key == "otel.status_code" || tag.Key == "status.code" {
						spanStatus = fmt.Sprintf("%v", tag.Value)
					}
					if tag.Key == "error" && fmt.Sprintf("%v", tag.Value) == "true" {
						spanStatus = "ERROR"
					}
				}

				durStr := formatDuration(rootSpan.Duration)
				ts := time.UnixMicro(rootSpan.StartTime).Format(time.RFC3339)

				table.Rows = append(table.Rows, []string{
					trace.TraceID,
					serviceName,
					rootSpan.OperationName,
					durStr,
					spanStatus,
					ts,
				})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "Filter by service name (e.g., \"api-gateway\")")
	cmd.Flags().StringVar(&env, "env", "", "Filter by environment (e.g., \"production\", \"staging\")")
	cmd.Flags().StringVar(&query, "query", "", "Filter by operation name (e.g., \"GET /api/users\")")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of traces to return")
	cmd.Flags().StringVar(&spanKind, "span-kind", "", "Filter by span kind (client, server, producer, consumer, internal)")
	cmd.Flags().StringVar(&status, "status", "", "Filter by span status: error, ok")
	cmd.Flags().StringVar(&minDuration, "min-duration", "", "Minimum trace duration (e.g., \"500ms\", \"1s\", \"100us\")")
	cmd.Flags().StringVar(&maxDuration, "max-duration", "", "Maximum trace duration (e.g., \"5s\", \"10s\")")
	cmd.Flags().StringArrayVar(&tags, "tags", nil, "Filter by span tag key=value (can be repeated, e.g., --tags \"http.method=POST\")")
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}

func formatDuration(microseconds int64) string {
	if microseconds < 1000 {
		return fmt.Sprintf("%dus", microseconds)
	}
	if microseconds < 1000000 {
		return fmt.Sprintf("%.1fms", float64(microseconds)/1000)
	}
	return fmt.Sprintf("%.2fs", float64(microseconds)/1000000)
}
