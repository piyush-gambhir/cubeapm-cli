package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestQueryLogs(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery, gotLimit string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		r.ParseForm()
		gotQuery = r.FormValue("query")
		gotLimit = r.FormValue("limit")
		// Return newline-delimited JSON
		fmt.Fprintln(w, `{"_msg":"error occurred","_time":"2024-01-15T10:00:00Z","_stream":"{host=\"web-1\"}"}`)
		fmt.Fprintln(w, `{"_msg":"another error","_time":"2024-01-15T10:01:00Z","_stream":"{host=\"web-2\"}"}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	entries, err := c.QueryLogs("error", from, to, 100)
	if err != nil {
		t.Fatalf("QueryLogs() error = %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/logs/select/logsql/query" {
		t.Errorf("path = %q, want /api/logs/select/logsql/query", gotPath)
	}
	if gotQuery != "error" {
		t.Errorf("query = %q, want %q", gotQuery, "error")
	}
	if gotLimit != "100" {
		t.Errorf("limit = %q, want %q", gotLimit, "100")
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].Message != "error occurred" {
		t.Errorf("entries[0].Message = %q, want %q", entries[0].Message, "error occurred")
	}
	if entries[0].Time != "2024-01-15T10:00:00Z" {
		t.Errorf("entries[0].Time = %q, want %q", entries[0].Time, "2024-01-15T10:00:00Z")
	}
}

func TestQueryLogs_WithLimit(t *testing.T) {
	var gotLimit string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotLimit = r.FormValue("limit")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.QueryLogs("*", time.Time{}, time.Time{}, 50)
	if err != nil {
		t.Fatalf("QueryLogs() error = %v", err)
	}

	if gotLimit != "50" {
		t.Errorf("limit = %q, want %q", gotLimit, "50")
	}
}

func TestGetHits(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery, gotStep string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = r.URL.Query().Get("query")
		gotStep = r.URL.Query().Get("step")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"hits": []map[string]interface{}{
				{
					"timestamps": []string{"2024-01-15T10:00:00Z", "2024-01-15T11:00:00Z"},
					"values":     []int64{100, 150},
					"fields":     map[string]string{},
				},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	result, err := c.GetLogHits("error", from, to, "1h")
	if err != nil {
		t.Fatalf("GetLogHits() error = %v", err)
	}

	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/logs/select/logsql/hits" {
		t.Errorf("path = %q, want /api/logs/select/logsql/hits", gotPath)
	}
	if gotQuery != "error" {
		t.Errorf("query = %q, want %q", gotQuery, "error")
	}
	if gotStep != "1h" {
		t.Errorf("step = %q, want %q", gotStep, "1h")
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Hits) != 1 {
		t.Errorf("got %d hit buckets, want 1", len(result.Hits))
	}
}

func TestGetStats(t *testing.T) {
	var gotPath, gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		// Stats query returns newline-delimited JSON
		fmt.Fprintln(w, `{"service":"api-gateway","count":"100"}`)
		fmt.Fprintln(w, `{"service":"payments","count":"50"}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	result, err := c.GetLogStats("error | stats count() by (service)", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetLogStats() error = %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/logs/select/logsql/stats_query" {
		t.Errorf("path = %q, want /api/logs/select/logsql/stats_query", gotPath)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(result.Rows))
	}
	if result.Rows[0].Fields["service"] != "api-gateway" {
		t.Errorf("rows[0].Fields[service] = %q, want %q", result.Rows[0].Fields["service"], "api-gateway")
	}
}

func TestGetStreams(t *testing.T) {
	var gotPath, gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		fmt.Fprintln(w, `{"value":"{host=\"web-1\"}","hits":500}`)
		fmt.Fprintln(w, `{"value":"{host=\"web-2\"}","hits":300}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	streams, err := c.GetLogStreams("", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetLogStreams() error = %v", err)
	}

	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/logs/select/logsql/streams" {
		t.Errorf("path = %q, want /api/logs/select/logsql/streams", gotPath)
	}
	if len(streams) != 2 {
		t.Fatalf("got %d streams, want 2", len(streams))
	}
	if streams[0].Stream != `{host="web-1"}` {
		t.Errorf("streams[0].Stream = %q, want %q", streams[0].Stream, `{host="web-1"}`)
	}
	if streams[0].Entries != 500 {
		t.Errorf("streams[0].Entries = %d, want %d", streams[0].Entries, 500)
	}
}

func TestGetFieldNames(t *testing.T) {
	var gotPath, gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		fmt.Fprintln(w, `{"value":"_msg","hits":1000}`)
		fmt.Fprintln(w, `{"value":"level","hits":800}`)
		fmt.Fprintln(w, `{"value":"service","hits":600}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	fields, err := c.GetLogFieldNames("", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetLogFieldNames() error = %v", err)
	}

	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/logs/select/logsql/field_names" {
		t.Errorf("path = %q, want /api/logs/select/logsql/field_names", gotPath)
	}
	if len(fields) != 3 {
		t.Fatalf("got %d fields, want 3", len(fields))
	}
	if fields[0].Name != "_msg" {
		t.Errorf("fields[0].Name = %q, want %q", fields[0].Name, "_msg")
	}
	if fields[0].Hits != 1000 {
		t.Errorf("fields[0].Hits = %d, want %d", fields[0].Hits, 1000)
	}
}

func TestGetFieldValues(t *testing.T) {
	var gotPath, gotMethod string
	var gotField string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotField = r.URL.Query().Get("field")
		fmt.Fprintln(w, `{"value":"error","hits":500}`)
		fmt.Fprintln(w, `{"value":"info","hits":300}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	values, err := c.GetLogFieldValues("level", "", time.Time{}, time.Time{}, 10)
	if err != nil {
		t.Fatalf("GetLogFieldValues() error = %v", err)
	}

	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/logs/select/logsql/field_values" {
		t.Errorf("path = %q, want /api/logs/select/logsql/field_values", gotPath)
	}
	if gotField != "level" {
		t.Errorf("field = %q, want %q", gotField, "level")
	}
	if len(values) != 2 {
		t.Fatalf("got %d values, want 2", len(values))
	}
	if values[0].Value != "error" {
		t.Errorf("values[0].Value = %q, want %q", values[0].Value, "error")
	}
}

func TestQueryLogs_EmptyResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty body (no log entries)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	entries, err := c.QueryLogs("nonexistent", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Fatalf("QueryLogs() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}

func TestQueryLogs_StreamingResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return multiple newline-delimited JSON entries
		for i := 0; i < 5; i++ {
			fmt.Fprintf(w, `{"_msg":"log entry %d","_time":"2024-01-15T10:%02d:00Z","_stream":"{host=\"web-1\"}","level":"info"}`, i, i)
			fmt.Fprintln(w)
		}
	}))
	defer ts.Close()

	c := newTestClient(ts)
	entries, err := c.QueryLogs("*", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Fatalf("QueryLogs() error = %v", err)
	}

	if len(entries) != 5 {
		t.Fatalf("got %d entries, want 5", len(entries))
	}

	// Verify streaming parsed all entries including extra fields
	if entries[0].Message != "log entry 0" {
		t.Errorf("entries[0].Message = %q, want %q", entries[0].Message, "log entry 0")
	}
	if entries[0].Fields["level"] != "info" {
		t.Errorf("entries[0].Fields[level] = %q, want %q", entries[0].Fields["level"], "info")
	}
}

func TestGetHits_ResponseFormat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"hits": []map[string]interface{}{
				{
					"timestamps": []string{"2024-01-15T10:00:00Z", "2024-01-15T11:00:00Z", "2024-01-15T12:00:00Z"},
					"values":     []int64{100, 150, 200},
					"fields":     map[string]string{},
				},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	result, err := c.GetLogHits("error", time.Time{}, time.Time{}, "1h")
	if err != nil {
		t.Fatalf("GetLogHits() error = %v", err)
	}

	if len(result.Hits) != 1 {
		t.Fatalf("got %d hit groups, want 1", len(result.Hits))
	}
	hit := result.Hits[0]
	if len(hit.Timestamps) != 3 {
		t.Errorf("got %d timestamps, want 3", len(hit.Timestamps))
	}
	if len(hit.Values) != 3 {
		t.Errorf("got %d values, want 3", len(hit.Values))
	}
}
