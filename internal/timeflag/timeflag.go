package timeflag

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ParseTime parses a time string in various formats:
//   - RFC3339 (2024-01-15T10:30:00Z)
//   - Unix timestamp (1705312200)
//   - Relative time (now, -1h, now-30m, -2d)
func ParseTime(input string) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// Handle "now"
	if strings.EqualFold(input, "now") {
		return time.Now(), nil
	}

	// Handle relative time: -1h, -30m, now-1h, etc.
	if strings.HasPrefix(input, "-") || strings.HasPrefix(strings.ToLower(input), "now-") {
		durStr := input
		if strings.HasPrefix(strings.ToLower(durStr), "now") {
			durStr = durStr[3:] // strip "now", leaving "-1h"
		}
		// durStr is now like "-1h", "-30m", "-2d"
		d, err := ParseDuration(strings.TrimPrefix(durStr, "-"))
		if err != nil {
			return time.Time{}, fmt.Errorf("parsing relative time %q: %w", input, err)
		}
		return time.Now().Add(-d), nil
	}

	// Try RFC3339
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t, nil
	}

	// Try RFC3339Nano
	if t, err := time.Parse(time.RFC3339Nano, input); err == nil {
		return t, nil
	}

	// Try date only
	if t, err := time.Parse("2006-01-02", input); err == nil {
		return t, nil
	}

	// Try date+time without timezone
	if t, err := time.Parse("2006-01-02T15:04:05", input); err == nil {
		return t, nil
	}

	// Try Unix timestamp
	if ts, err := strconv.ParseInt(input, 10, 64); err == nil {
		if ts > 1e12 {
			// milliseconds
			return time.UnixMilli(ts), nil
		}
		return time.Unix(ts, 0), nil
	}

	// Try Unix float timestamp
	if ts, err := strconv.ParseFloat(input, 64); err == nil {
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time %q: use RFC3339, Unix timestamp, or relative format (-1h, now-30m)", input)
}

// ParseDuration parses a duration string that supports d/h/m/s suffixes.
// Go's time.ParseDuration does not support 'd' (days), so we handle that.
// Examples: "1h", "30m", "2d", "1d12h", "90s", "1h30m"
func ParseDuration(input string) (time.Duration, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Check if it contains 'd' for days
	if strings.Contains(input, "d") {
		return parseDurationWithDays(input)
	}

	// Fall back to Go's built-in parser
	d, err := time.ParseDuration(input)
	if err != nil {
		return 0, fmt.Errorf("parsing duration %q: %w", input, err)
	}
	return d, nil
}

func parseDurationWithDays(input string) (time.Duration, error) {
	var total time.Duration
	remaining := input

	for remaining != "" {
		// Find the next number
		i := 0
		for i < len(remaining) && (remaining[i] >= '0' && remaining[i] <= '9' || remaining[i] == '.') {
			i++
		}
		if i == 0 {
			return 0, fmt.Errorf("invalid duration %q: expected number", input)
		}
		if i >= len(remaining) {
			return 0, fmt.Errorf("invalid duration %q: missing unit", input)
		}

		numStr := remaining[:i]
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", input, err)
		}

		unit := remaining[i]
		switch unit {
		case 'd':
			total += time.Duration(num * float64(24*time.Hour))
		case 'h':
			total += time.Duration(num * float64(time.Hour))
		case 'm':
			total += time.Duration(num * float64(time.Minute))
		case 's':
			total += time.Duration(num * float64(time.Second))
		default:
			return 0, fmt.Errorf("invalid duration %q: unknown unit %q", input, string(unit))
		}

		remaining = remaining[i+1:]
	}

	return total, nil
}

// ResolveTimeRange resolves --from, --to, and --last flags into a concrete
// time range. If none are provided, defaults to --last 1h.
func ResolveTimeRange(from, to, last string) (start, end time.Time, err error) {
	now := time.Now()

	// If --last is specified, it takes precedence as a convenience flag
	if last != "" {
		d, err := ParseDuration(last)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --last: %w", err)
		}
		return now.Add(-d), now, nil
	}

	// If both --from and --to are set
	if from != "" && to != "" {
		start, err = ParseTime(from)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --from: %w", err)
		}
		end, err = ParseTime(to)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --to: %w", err)
		}
		return start, end, nil
	}

	// If only --from, default --to to now
	if from != "" {
		start, err = ParseTime(from)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --from: %w", err)
		}
		return start, now, nil
	}

	// If only --to, default --from to 1h before --to
	if to != "" {
		end, err = ParseTime(to)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --to: %w", err)
		}
		return end.Add(-1 * time.Hour), end, nil
	}

	// Default: last 1 hour
	return now.Add(-1 * time.Hour), now, nil
}

// AddTimeFlags adds --from, --to, and --last flags to a cobra command.
func AddTimeFlags(cmd *cobra.Command, from, to, last *string) {
	cmd.Flags().StringVar(from, "from", "", "Start time (RFC3339, Unix timestamp, or relative like -1h)")
	cmd.Flags().StringVar(to, "to", "", "End time (RFC3339, Unix timestamp, or relative)")
	cmd.Flags().StringVar(last, "last", "", "Relative duration from now (e.g., 1h, 30m, 2d)")
}
