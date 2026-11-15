package timeflag

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func ParseTime(input string) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	if strings.EqualFold(input, "now") {
		return time.Now(), nil
	}

	if strings.HasPrefix(input, "-") || strings.HasPrefix(strings.ToLower(input), "now-") {
		durStr := input
		if strings.HasPrefix(strings.ToLower(durStr), "now") {
			durStr = durStr[3:]
		}
		d, err := ParseDuration(strings.TrimPrefix(durStr, "-"))
		if err != nil {
			return time.Time{}, fmt.Errorf("parsing relative time %q: %w", input, err)
		}
		return time.Now().Add(-d), nil
	}

	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t, nil
	}

	if t, err := time.Parse(time.RFC3339Nano, input); err == nil {
		return t, nil
	}

	if t, err := time.ParseInLocation("2006-01-02", input, time.Local); err == nil {
		return t, nil
	}

	if t, err := time.ParseInLocation("2006-01-02T15:04:05", input, time.Local); err == nil {
		return t, nil
	}

	if ts, err := strconv.ParseInt(input, 10, 64); err == nil {
		if ts >= int64(1e12) {
			return time.UnixMilli(ts), nil
		}
		return time.Unix(ts, 0), nil
	}

	if ts, err := strconv.ParseFloat(input, 64); err == nil {
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time %q: use RFC3339, Unix timestamp, or relative format (-1h, now-30m)", input)
}

func ParseDuration(input string) (time.Duration, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	if err := validateDurationFormat(input); err != nil {
		return 0, err
	}

	if strings.Contains(input, "d") {
		return parseDurationWithDays(input)
	}

	d, err := time.ParseDuration(input)
	if err != nil {
		return 0, fmt.Errorf("parsing duration %q: %w", input, err)
	}
	return d, nil
}

func validateDurationFormat(input string) error {
	remaining := input
	found := false
	for remaining != "" {
		i := 0
		for i < len(remaining) && (remaining[i] >= '0' && remaining[i] <= '9' || remaining[i] == '.') {
			i++
		}
		if i == 0 {
			return fmt.Errorf("invalid duration %q: unexpected character %q", input, string(remaining[0]))
		}
		if i >= len(remaining) {
			return fmt.Errorf("invalid duration %q: missing unit suffix", input)
		}

		rest := remaining[i:]
		matched := false
		for _, unit := range []string{"ms", "us", "ns", "d", "h", "m", "s"} {
			if strings.HasPrefix(rest, unit) {
				remaining = rest[len(unit):]
				matched = true
				found = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("invalid duration %q: unknown unit starting at %q", input, rest)
		}
	}
	if !found {
		return fmt.Errorf("invalid duration %q", input)
	}
	return nil
}

func parseDurationWithDays(input string) (time.Duration, error) {
	var total time.Duration
	remaining := input

	for remaining != "" {
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

func ResolveTimeRange(from, to, last string) (start, end time.Time, err error) {
	now := time.Now()

	if last != "" {
		d, err := ParseDuration(last)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --last: %w", err)
		}
		return now.Add(-d), now, nil
	}

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

	if from != "" {
		start, err = ParseTime(from)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --from: %w", err)
		}
		return start, now, nil
	}

	if to != "" {
		end, err = ParseTime(to)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("parsing --to: %w", err)
		}
		return end.Add(-1 * time.Hour), end, nil
	}

	return now.Add(-1 * time.Hour), now, nil
}

func AddTimeFlags(cmd *cobra.Command, from, to, last *string) {
	cmd.Flags().StringVar(from, "from", "", "Start time (RFC3339, Unix timestamp, or relative like -1h; timestamps without a timezone are local time)")
	cmd.Flags().StringVar(to, "to", "", "End time (RFC3339, Unix timestamp, or relative; timestamps without a timezone are local time)")
	cmd.Flags().StringVar(last, "last", "", "Relative duration from now (e.g., 1h, 30m, 2d)")
}
