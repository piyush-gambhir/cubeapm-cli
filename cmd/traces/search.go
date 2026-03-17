package traces

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
)

func newSearchCmd() *cobra.Command {
	var (
		service  string
		env      string
		query    string
		from     string
		to       string
		last     string
		limit    int
		spanKind string
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for traces",
		Long:  "Search for traces matching the given criteria. Returns a summary of matching traces.",
		RunE: func(cmd *cobra.Command, args []string) error {
			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}

			results, err := cmdutil.APIClient.SearchTraces(service, env, query, start, end, limit, spanKind)
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
				status := "OK"
				for _, tag := range rootSpan.Tags {
					if tag.Key == "otel.status_code" || tag.Key == "status.code" {
						status = fmt.Sprintf("%v", tag.Value)
					}
					if tag.Key == "error" && fmt.Sprintf("%v", tag.Value) == "true" {
						status = "ERROR"
					}
				}

				durStr := formatDuration(rootSpan.Duration)
				ts := time.UnixMicro(rootSpan.StartTime).Format(time.RFC3339)

				table.Rows = append(table.Rows, []string{
					trace.TraceID,
					serviceName,
					rootSpan.OperationName,
					durStr,
					status,
					ts,
				})
			}

			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	cmd.Flags().StringVar(&service, "service", "", "Filter by service name")
	cmd.Flags().StringVar(&env, "env", "", "Filter by environment")
	cmd.Flags().StringVar(&query, "query", "", "Filter by operation name")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of traces to return")
	cmd.Flags().StringVar(&spanKind, "span-kind", "", "Filter by span kind (client, server, producer, consumer, internal)")
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
