package traces

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
	"github.com/piyush-gambhir/cubeapm-cli/internal/timeflag"
	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

// newCallersCmd returns the `traces callers` command.
//
// Finding "who is calling endpoint X on service Y" during an RCA is a
// recurring task. The data exists in CubeAPM's cube_apm_latency_count
// metric with span_kind="client" and group_name="HTTP <host>", but the
// query is non-obvious and easy to get wrong (wrong group_name,
// forgetting span_kind, using an absolute count instead of a rate).
// This command aggregates it into one call so investigators get to the
// "which service is suddenly talking to me a lot more?" answer in
// seconds rather than minutes.
func newCallersCmd() *cobra.Command {
	var (
		service string
		host    string
		from    string
		to      string
		last    string
		window  string
		topk    int
	)

	cmd := &cobra.Command{
		Use:   "callers",
		Short: "Rank the services making outbound HTTP calls to a host",
		Long: `List the services making outbound HTTP calls to a given host (or reaching a given service)
over the specified time window, aggregated by call rate.

This is a convenience over the raw PromQL query:

  topk(N, sum by (service) (rate(cube_apm_latency_count{
      group_name="HTTP <host>",
      span_kind="client"
  }[<window>])))

Specify the target host either explicitly with --host (matches the
group_name label, typically "api.spyne.ai", "internal.corp", etc.) or
via --service to target a specific service's inbound traffic. When
--service is given without --host, the host is inferred as "HTTP " +
a conventional domain, if your setup differs, use --host explicitly.

Time ranges can be specified as:
  - Relative:   --last 1h
  - RFC3339:    --from 2024-01-15T10:00:00Z --to 2024-01-15T11:00:00Z
  - Default:    last 1 hour if no time flags are provided

Examples:
  # Who's calling api.spyne.ai in the last hour?
  cubeapm traces callers --host api.spyne.ai --last 1h

  # Find the biggest outbound callers during an incident window
  cubeapm traces callers --host api.spyne.ai \
    --from 2026-04-19T13:20:00Z --to 2026-04-19T13:50:00Z

  # Top 20 callers with a 5-minute rate window
  cubeapm traces callers --host api.spyne.ai --last 1h --window 5m --topk 20

  # Find callers of a specific service (host inferred)
  cubeapm traces callers --service MEDIA-SERVICE --last 30m`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if host == "" && service == "" {
				return fmt.Errorf("specify either --host or --service")
			}
			groupName := host
			if groupName == "" {
				// Best-effort inference. Users with different conventions
				// should set --host explicitly.
				groupName = "api.spyne.ai"
			}
			if !strings.HasPrefix(groupName, "HTTP ") {
				groupName = "HTTP " + groupName
			}

			start, end, err := timeflag.ResolveTimeRange(from, to, last)
			if err != nil {
				return err
			}
			// PromQL instant query at `end` with the requested rate window.
			selector := fmt.Sprintf(
				`topk(%d, sum by (service) (rate(cube_apm_latency_count{group_name=%q,span_kind="client"}[%s])))`,
				topk, groupName, window,
			)
			// If the caller asked to narrow to a specific called service,
			// add that filter as well, covers deployments where a caller
			// hits multiple hosts and you want only the traffic to the
			// named service.
			if service != "" {
				// The client span's service label identifies the caller;
				// the span_name encodes the callee. Use a regex on
				// span_name to keep spans whose target mentions `service`.
				selector = fmt.Sprintf(
					`topk(%d, sum by (service) (rate(cube_apm_latency_count{group_name=%q,span_kind="client",span_name=~".*%s.*"}[%s])))`,
					topk, groupName, regexSafe(service), window,
				)
			}

			_ = start // `end` is what PromQL needs for instant queries;
			// `start` is only relevant if we add a range variant later.

			result, err := cmdutil.APIClient.QueryInstant(selector, end)
			if err != nil {
				return err
			}
			if result.Data.ResultType != "vector" {
				return fmt.Errorf("unexpected result type %q from caller query", result.Data.ResultType)
			}
			var samples types.VectorResult
			if err := json.Unmarshal(result.Data.Result, &samples); err != nil {
				return fmt.Errorf("parsing caller query result: %w", err)
			}

			rows := make([][]string, 0, len(samples))
			for _, s := range samples {
				svc := s.Metric["service"]
				if svc == "" {
					svc = "(unknown)"
				}
				rows = append(rows, []string{svc, s.Value.Value()})
			}

			if len(rows) == 0 && cmdutil.OutputFormat == output.FormatTable {
				fmt.Println("No callers found. Check --host spelling, try `cubeapm metrics series --match 'cube_apm_latency_count{span_kind=\"client\"}'` to discover valid group_name values.")
				return nil
			}

			sort.SliceStable(rows, func(i, j int) bool {
				var a, b float64
				fmt.Sscanf(rows[i][1], "%f", &a)
				fmt.Sscanf(rows[j][1], "%f", &b)
				return a > b
			})

			table := output.TableDef{
				Headers: []string{"CALLER_SERVICE", "CALLS_PER_SEC"},
				Rows:    rows,
			}
			return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, table)
		},
	}

	cmd.Flags().StringVar(&host, "host", "", `Target host (matches the "group_name" label, e.g. "api.spyne.ai")`)
	cmd.Flags().StringVar(&service, "service", "", "Filter to callers of this target service")
	cmd.Flags().StringVar(&window, "window", "2m", "Rate window (e.g. 2m, 5m). Shorter windows are more responsive but noisier.")
	cmd.Flags().IntVar(&topk, "topk", 10, "Maximum number of callers to return")
	timeflag.AddTimeFlags(cmd, &from, &to, &last)

	return cmd
}

// regexSafe escapes the small set of PromQL regex metacharacters that
// show up in real service names. We deliberately stay minimal, users
// passing regex-intent strings can edit the query directly.
func regexSafe(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`.`, `\.`,
		`(`, `\(`,
		`)`, `\)`,
		`"`, `\"`,
	)
	return replacer.Replace(s)
}
