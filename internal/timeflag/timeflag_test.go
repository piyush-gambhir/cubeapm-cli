package timeflag

import (
	"math"
	"testing"
	"time"
)

func TestParseTime_RFC3339(t *testing.T) {
	result, err := ParseTime("2024-01-15T10:00:00Z")
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}
	expected := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("ParseTime() = %v, want %v", result, expected)
	}
}

func TestParseTime_RFC3339_WithTZ(t *testing.T) {
	result, err := ParseTime("2024-01-15T10:00:00+05:30")
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}

	ist := time.FixedZone("IST", 5*3600+30*60)
	expected := time.Date(2024, 1, 15, 10, 0, 0, 0, ist)
	if !result.Equal(expected) {
		t.Errorf("ParseTime() = %v, want %v", result, expected)
	}
}

func TestParseTime_UnixSeconds(t *testing.T) {
	result, err := ParseTime("1705312800")
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}
	expected := time.Unix(1705312800, 0)
	if !result.Equal(expected) {
		t.Errorf("ParseTime() = %v, want %v", result, expected)
	}
}

func TestParseTime_UnixMillis(t *testing.T) {
	result, err := ParseTime("1705312800000")
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}
	expected := time.UnixMilli(1705312800000)
	if !result.Equal(expected) {
		t.Errorf("ParseTime() = %v, want %v", result, expected)
	}
}

func TestParseTime_Relative_Hours(t *testing.T) {
	before := time.Now()
	result, err := ParseTime("-1h")
	after := time.Now()
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}

	expectedLow := before.Add(-1 * time.Hour)
	expectedHigh := after.Add(-1 * time.Hour)
	if result.Before(expectedLow) || result.After(expectedHigh) {
		t.Errorf("ParseTime(-1h) = %v, expected between %v and %v", result, expectedLow, expectedHigh)
	}
}

func TestParseTime_Relative_Minutes(t *testing.T) {
	before := time.Now()
	result, err := ParseTime("-30m")
	after := time.Now()
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}

	expectedLow := before.Add(-30 * time.Minute)
	expectedHigh := after.Add(-30 * time.Minute)
	if result.Before(expectedLow) || result.After(expectedHigh) {
		t.Errorf("ParseTime(-30m) = %v, expected between %v and %v", result, expectedLow, expectedHigh)
	}
}

func TestParseTime_Relative_Days(t *testing.T) {
	before := time.Now()
	result, err := ParseTime("-2d")
	after := time.Now()
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}

	expectedLow := before.Add(-2 * 24 * time.Hour)
	expectedHigh := after.Add(-2 * 24 * time.Hour)
	if result.Before(expectedLow) || result.After(expectedHigh) {
		t.Errorf("ParseTime(-2d) = %v, expected between %v and %v", result, expectedLow, expectedHigh)
	}
}

func TestParseTime_Now(t *testing.T) {
	before := time.Now()
	result, err := ParseTime("now")
	after := time.Now()
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}

	if result.Before(before) || result.After(after) {
		t.Errorf("ParseTime(now) = %v, expected between %v and %v", result, before, after)
	}
}

func TestParseTime_Invalid(t *testing.T) {
	_, err := ParseTime("foobar")
	if err == nil {
		t.Fatal("ParseTime(foobar) expected error, got nil")
	}
}

func TestResolveTimeRange_LastFlag(t *testing.T) {
	before := time.Now()
	start, end, err := ResolveTimeRange("", "", "1h")
	after := time.Now()
	if err != nil {
		t.Fatalf("ResolveTimeRange() error = %v", err)
	}

	// end should be approximately now
	if end.Before(before) || end.After(after) {
		t.Errorf("end = %v, expected between %v and %v", end, before, after)
	}

	// start should be approximately 1 hour before now
	expectedStart := before.Add(-1 * time.Hour)
	diff := math.Abs(float64(start.Sub(expectedStart)))
	if diff > float64(2*time.Second) {
		t.Errorf("start = %v, expected approximately %v (diff: %v)", start, expectedStart, time.Duration(diff))
	}
}

func TestResolveTimeRange_ExplicitRange(t *testing.T) {
	start, end, err := ResolveTimeRange("2024-01-15T10:00:00Z", "2024-01-15T12:00:00Z", "")
	if err != nil {
		t.Fatalf("ResolveTimeRange() error = %v", err)
	}

	expectedStart := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	if !start.Equal(expectedStart) {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}
	if !end.Equal(expectedEnd) {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}
}

func TestResolveTimeRange_OnlyFrom(t *testing.T) {
	before := time.Now()
	start, end, err := ResolveTimeRange("2024-01-15T10:00:00Z", "", "")
	after := time.Now()
	if err != nil {
		t.Fatalf("ResolveTimeRange() error = %v", err)
	}

	expectedStart := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	if !start.Equal(expectedStart) {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}

	// end should default to now
	if end.Before(before) || end.After(after) {
		t.Errorf("end = %v, expected between %v and %v", end, before, after)
	}
}

func TestResolveTimeRange_OnlyTo(t *testing.T) {
	start, end, err := ResolveTimeRange("", "2024-01-15T12:00:00Z", "")
	if err != nil {
		t.Fatalf("ResolveTimeRange() error = %v", err)
	}

	expectedEnd := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	if !end.Equal(expectedEnd) {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}

	// start should be 1 hour before end
	expectedStart := expectedEnd.Add(-1 * time.Hour)
	if !start.Equal(expectedStart) {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}
}

func TestResolveTimeRange_Default(t *testing.T) {
	before := time.Now()
	start, end, err := ResolveTimeRange("", "", "")
	after := time.Now()
	if err != nil {
		t.Fatalf("ResolveTimeRange() error = %v", err)
	}

	// end should be approximately now
	if end.Before(before) || end.After(after) {
		t.Errorf("end = %v, expected between %v and %v", end, before, after)
	}

	// start should be approximately 1 hour before now (default)
	expectedStart := before.Add(-1 * time.Hour)
	diff := math.Abs(float64(start.Sub(expectedStart)))
	if diff > float64(2*time.Second) {
		t.Errorf("start = %v, expected approximately %v", start, expectedStart)
	}
}

// --- Boundary tests for seconds vs milliseconds ---

func TestParseTime_BoundaryBelowE12_Seconds(t *testing.T) {
	// 999999999999 is just below 1e12, should be treated as seconds.
	result, err := ParseTime("999999999999")
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}
	expected := time.Unix(999999999999, 0)
	if !result.Equal(expected) {
		t.Errorf("ParseTime(999999999999) = %v, want %v (seconds)", result, expected)
	}
}

func TestParseTime_BoundaryAtE12_Millis(t *testing.T) {
	// 1000000000000 is exactly 1e12, should be treated as milliseconds.
	result, err := ParseTime("1000000000000")
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}
	expected := time.UnixMilli(1000000000000)
	if !result.Equal(expected) {
		t.Errorf("ParseTime(1000000000000) = %v, want %v (millis)", result, expected)
	}
}

func TestParseTime_BoundaryAt1e10_Seconds(t *testing.T) {
	// 10000000000 (1e10) should still be treated as seconds.
	result, err := ParseTime("10000000000")
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}
	expected := time.Unix(10000000000, 0)
	if !result.Equal(expected) {
		t.Errorf("ParseTime(10000000000) = %v, want %v (seconds)", result, expected)
	}
}

// --- Duration validation tests ---

func TestParseDuration_InvalidPrefix(t *testing.T) {
	_, err := ParseDuration("1x2h")
	if err == nil {
		t.Fatal("ParseDuration(1x2h) expected error, got nil")
	}
}

func TestParseDuration_InvalidSuffix(t *testing.T) {
	_, err := ParseDuration("abc")
	if err == nil {
		t.Fatal("ParseDuration(abc) expected error, got nil")
	}
}

func TestParseDuration_ValidCompound(t *testing.T) {
	d, err := ParseDuration("1h30m")
	if err != nil {
		t.Fatalf("ParseDuration(1h30m) error = %v", err)
	}
	expected := 1*time.Hour + 30*time.Minute
	if d != expected {
		t.Errorf("ParseDuration(1h30m) = %v, want %v", d, expected)
	}
}

func TestParseDuration_ValidDaysAndHours(t *testing.T) {
	d, err := ParseDuration("2d12h")
	if err != nil {
		t.Fatalf("ParseDuration(2d12h) error = %v", err)
	}
	expected := 2*24*time.Hour + 12*time.Hour
	if d != expected {
		t.Errorf("ParseDuration(2d12h) = %v, want %v", d, expected)
	}
}

func TestParseDuration_TrailingGarbage(t *testing.T) {
	_, err := ParseDuration("1hxyz")
	if err == nil {
		t.Fatal("ParseDuration(1hxyz) expected error, got nil")
	}
}

func TestParseDuration_ValidMilliseconds(t *testing.T) {
	d, err := ParseDuration("500ms")
	if err != nil {
		t.Fatalf("ParseDuration(500ms) error = %v", err)
	}
	expected := 500 * time.Millisecond
	if d != expected {
		t.Errorf("ParseDuration(500ms) = %v, want %v", d, expected)
	}
}

func TestResolveTimeRange_LastOverridesFromTo(t *testing.T) {
	before := time.Now()
	start, end, err := ResolveTimeRange("2024-01-01T00:00:00Z", "2024-12-31T23:59:59Z", "30m")
	after := time.Now()
	if err != nil {
		t.Fatalf("ResolveTimeRange() error = %v", err)
	}

	// --last should take precedence: end = now, start = now - 30m
	if end.Before(before) || end.After(after) {
		t.Errorf("end = %v, expected between %v and %v (--last should override --to)", end, before, after)
	}

	expectedStart := before.Add(-30 * time.Minute)
	diff := math.Abs(float64(start.Sub(expectedStart)))
	if diff > float64(2*time.Second) {
		t.Errorf("start = %v, expected approximately %v (--last should override --from)", start, expectedStart)
	}
}
