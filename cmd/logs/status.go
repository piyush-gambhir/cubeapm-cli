package logs

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
)

// LogsStatus is the structured shape emitted by `cubeapm logs status -o json`.
type LogsStatus struct {
	Query                 string `json:"query"`
	LookbackDays          int    `json:"lookbackDays"`
	HasLogs               bool   `json:"hasLogs"`
	EarliestNonZeroBucket string `json:"earliestNonZeroBucket,omitempty"`
	LatestNonZeroBucket   string `json:"latestNonZeroBucket,omitempty"`
	RetentionHours        int    `json:"retentionHours,omitempty"`
	NonEmptyDays          int    `json:"nonEmptyDays"`
	Note                  string `json:"note,omitempty"`
}

func newStatusCmd() *cobra.Command {
	var (
		query    string
		lookback int
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Probe log retention (oldest and newest available log buckets)",
		Long: `Probe the CubeAPM logs backend to determine the effective retention window.

This subcommand exists because CubeAPM's logs API does not expose retention
configuration directly. Instead, this command samples 24-hour buckets from the
last N days (default 30) and reports:

  - earliestNonZeroBucket — the oldest day with at least one log entry
  - latestNonZeroBucket   — the most recent day with at least one log entry
  - retentionHours        — wall-clock hours from earliest bucket to now
  - nonEmptyDays          — how many of the sampled days had any logs

Use it before running an RCA on an incident more than a day or two old: if the
earliest non-zero bucket is more recent than the incident window, the logs for
that window have already been aged out and 'logs query' will return empty.

The --query flag accepts a LogsQL expression; if omitted, defaults to '*'
(all logs). Use --query 'service.name:X' to probe retention for a specific
service.

Examples:
  # How far back do logs go right now?
  cubeapm logs status

  # Same, but for a specific service
  cubeapm logs status --query 'service.name:MEDIA-SERVICE'

  # Look further back (60 days)
  cubeapm logs status --lookback 60

  # Structured output for scripting
  cubeapm logs status -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if query == "" {
				query = "*"
			}
			if lookback <= 0 {
				lookback = 30
			}

			now := time.Now().UTC()
			from := now.Add(-time.Duration(lookback) * 24 * time.Hour)

			result, err := cmdutil.APIClient.GetLogHits(query, from, now, "24h")
			if err != nil {
				return fmt.Errorf("probing log hits: %w", err)
			}

			var (
				earliest, latest string
				nonEmpty         int
			)
			for _, hit := range result.Hits {
				for i, ts := range hit.Timestamps {
					if i >= len(hit.Values) || hit.Values[i] == 0 {
						continue
					}
					nonEmpty++
					if earliest == "" || ts < earliest {
						earliest = ts
					}
					if latest == "" || ts > latest {
						latest = ts
					}
				}
			}

			status := LogsStatus{
				Query:        query,
				LookbackDays: lookback,
				HasLogs:      earliest != "",
				NonEmptyDays: nonEmpty,
			}

			if status.HasLogs {
				status.EarliestNonZeroBucket = earliest
				status.LatestNonZeroBucket = latest
				if t, perr := time.Parse(time.RFC3339, earliest); perr == nil {
					status.RetentionHours = int(now.Sub(t).Hours())
				}
				if status.RetentionHours > 0 && status.RetentionHours < 24*(lookback-1) {
					status.Note = fmt.Sprintf("oldest non-zero bucket is %d h ago, less than the %d-day lookback — likely the actual retention horizon", status.RetentionHours, lookback)
				}
			} else {
				status.Note = fmt.Sprintf("no logs found in the last %d days matching query %q — either logs are not flowing for this query, or retention is shorter than the smallest bucket (24 h)", lookback, query)
			}

			if cmdutil.OutputFormat != output.FormatTable {
				return output.Print(cmdutil.OutputFormat, cmdutil.Resolved.NoColor, status)
			}

			// Human-readable table.
			fmt.Println("CubeAPM logs retention probe")
			fmt.Printf("  Query:        %s\n", status.Query)
			fmt.Printf("  Lookback:     %d days\n", status.LookbackDays)
			fmt.Printf("  Non-empty days: %d\n", status.NonEmptyDays)
			if status.HasLogs {
				fmt.Printf("  Earliest:     %s\n", status.EarliestNonZeroBucket)
				fmt.Printf("  Latest:       %s\n", status.LatestNonZeroBucket)
				if status.RetentionHours > 0 {
					if status.RetentionHours < 72 {
						fmt.Printf("  Retention:    %d h (%.1f days)\n", status.RetentionHours, float64(status.RetentionHours)/24)
					} else {
						fmt.Printf("  Retention:    ~%d days\n", status.RetentionHours/24)
					}
				}
			} else {
				fmt.Println("  Earliest:     <none>")
			}
			if status.Note != "" {
				fmt.Println()
				fmt.Println(strings.TrimRight("Note: "+status.Note, " "))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "LogsQL query to probe (default: *)")
	cmd.Flags().IntVar(&lookback, "lookback", 30, "Lookback window in days for the probe")

	return cmd
}
