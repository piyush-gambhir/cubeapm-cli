package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func newTestClient(ts *httptest.Server) *Client {
	return &Client{
		queryBaseURL:  ts.URL,
		ingestBaseURL: ts.URL,
		adminBaseURL:  ts.URL,
		httpClient:    ts.Client(),
	}
}

func TestGetServices(t *testing.T) {
	var gotPath, gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []string{"api-gateway", "payments", "auth"},
			"total":  3,
			"errors": []interface{}{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	services, err := c.GetServices()
	if err != nil {
		t.Fatalf("GetServices() error = %v", err)
	}

	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/traces/api/v1/services" {
		t.Errorf("path = %q, want /api/traces/api/v1/services", gotPath)
	}
	if len(services) != 3 {
		t.Errorf("got %d services, want 3", len(services))
	}
	if services[0] != "api-gateway" {
		t.Errorf("services[0] = %q, want %q", services[0], "api-gateway")
	}
}

func TestGetOperations(t *testing.T) {
	var gotPath, gotMethod string
	var gotSpanKind string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotSpanKind = r.URL.Query().Get("spanKind")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{
				{"name": "GET /api/users", "spanKind": "server"},
				{"name": "POST /api/orders", "spanKind": "server"},
			},
			"errors": []interface{}{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	ops, err := c.GetOperations("api-gateway", "server")
	if err != nil {
		t.Fatalf("GetOperations() error = %v", err)
	}

	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/traces/api/v1/services/api-gateway/operations" {
		t.Errorf("path = %q, want /api/traces/api/v1/services/api-gateway/operations", gotPath)
	}
	if gotSpanKind != "server" {
		t.Errorf("spanKind param = %q, want %q", gotSpanKind, "server")
	}
	if len(ops) != 2 {
		t.Errorf("got %d operations, want 2", len(ops))
	}
}

func TestSearchTraces_Minimal(t *testing.T) {
	var gotPath string
	var gotService string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotService = r.URL.Query().Get("service")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []interface{}{},
			"errors": []interface{}{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	results, err := c.SearchTraces("api-gateway", nil, "", time.Time{}, time.Time{}, 0, "", "", "")
	if err != nil {
		t.Fatalf("SearchTraces() error = %v", err)
	}

	if gotPath != "/api/traces/api/v1/search" {
		t.Errorf("path = %q, want /api/traces/api/v1/search", gotPath)
	}
	if gotService != "api-gateway" {
		t.Errorf("service param = %q, want %q", gotService, "api-gateway")
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestSearchTraces_AllParams(t *testing.T) {
	var gotQuery url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []interface{}{},
			"errors": []interface{}{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	tags := map[string]string{"http.method": "GET"}

	_, err := c.SearchTraces("payments", tags, "POST /orders", from, to, 50, "server", "500ms", "2s")
	if err != nil {
		t.Fatalf("SearchTraces() error = %v", err)
	}

	if gotQuery.Get("service") != "payments" {
		t.Errorf("service = %q, want %q", gotQuery.Get("service"), "payments")
	}
	if gotQuery.Get("operation") != "POST /orders" {
		t.Errorf("operation = %q, want %q", gotQuery.Get("operation"), "POST /orders")
	}
	if gotQuery.Get("limit") != "50" {
		t.Errorf("limit = %q, want %q", gotQuery.Get("limit"), "50")
	}
	if gotQuery.Get("spanKind") != "server" {
		t.Errorf("spanKind = %q, want %q", gotQuery.Get("spanKind"), "server")
	}
	if gotQuery.Get("minDuration") != "500ms" {
		t.Errorf("minDuration = %q, want %q", gotQuery.Get("minDuration"), "500ms")
	}
	if gotQuery.Get("maxDuration") != "2s" {
		t.Errorf("maxDuration = %q, want %q", gotQuery.Get("maxDuration"), "2s")
	}
	if gotQuery.Get("start") == "" {
		t.Error("start param is empty, expected Unix micros")
	}
	if gotQuery.Get("end") == "" {
		t.Error("end param is empty, expected Unix micros")
	}
	if gotQuery.Get("tags") == "" {
		t.Error("tags param is empty, expected JSON")
	}
}

func TestSearchTraces_WithTags(t *testing.T) {
	var gotTags string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTags = r.URL.Query().Get("tags")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []interface{}{},
			"errors": []interface{}{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	tags := map[string]string{"http.method": "POST", "http.status_code": "500"}

	_, err := c.SearchTraces("api-gateway", tags, "", time.Time{}, time.Time{}, 0, "", "", "")
	if err != nil {
		t.Fatalf("SearchTraces() error = %v", err)
	}

	// Verify tags is valid JSON
	var parsed map[string]string
	if err := json.Unmarshal([]byte(gotTags), &parsed); err != nil {
		t.Fatalf("tags param is not valid JSON: %v", err)
	}
	if parsed["http.method"] != "POST" {
		t.Errorf("tags[http.method] = %q, want %q", parsed["http.method"], "POST")
	}
	if parsed["http.status_code"] != "500" {
		t.Errorf("tags[http.status_code] = %q, want %q", parsed["http.status_code"], "500")
	}
}

func TestSearchTraces_WithDuration(t *testing.T) {
	var gotMinDur, gotMaxDur string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMinDur = r.URL.Query().Get("minDuration")
		gotMaxDur = r.URL.Query().Get("maxDuration")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []interface{}{},
			"errors": []interface{}{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.SearchTraces("api-gateway", nil, "", time.Time{}, time.Time{}, 0, "", "100ms", "5s")
	if err != nil {
		t.Fatalf("SearchTraces() error = %v", err)
	}

	if gotMinDur != "100ms" {
		t.Errorf("minDuration = %q, want %q", gotMinDur, "100ms")
	}
	if gotMaxDur != "5s" {
		t.Errorf("maxDuration = %q, want %q", gotMaxDur, "5s")
	}
}

func TestGetTrace(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"traceID": "abc123def456",
					"spans": []map[string]interface{}{
						{
							"traceID":       "abc123def456",
							"spanID":        "span001",
							"operationName": "GET /api/users",
							"startTime":     1705312800000000,
							"duration":      150000,
							"processID":     "p1",
						},
					},
					"processes": map[string]interface{}{
						"p1": map[string]interface{}{
							"serviceName": "api-gateway",
						},
					},
				},
			},
			"errors": []interface{}{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	trace, err := c.GetTrace("abc123def456", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetTrace() error = %v", err)
	}

	if gotPath != "/api/traces/api/v1/traces/abc123def456" {
		t.Errorf("path = %q, want /api/traces/api/v1/traces/abc123def456", gotPath)
	}
	if trace.TraceID != "abc123def456" {
		t.Errorf("traceID = %q, want %q", trace.TraceID, "abc123def456")
	}
	if len(trace.Spans) != 1 {
		t.Errorf("got %d spans, want 1", len(trace.Spans))
	}
}

func TestGetDependencies(t *testing.T) {
	var gotPath string
	var gotEndTs, gotLookback string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotEndTs = r.URL.Query().Get("endTs")
		gotLookback = r.URL.Query().Get("lookback")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"parent": "api-gateway", "child": "payments", "callCount": 100},
				{"parent": "api-gateway", "child": "auth", "callCount": 50},
			},
			"errors": []interface{}{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	deps, err := c.GetDependencies(from, to)
	if err != nil {
		t.Fatalf("GetDependencies() error = %v", err)
	}

	if gotPath != "/api/traces/api/v1/dependencies" {
		t.Errorf("path = %q, want /api/traces/api/v1/dependencies", gotPath)
	}
	if gotEndTs == "" {
		t.Error("endTs param is empty")
	}
	if gotLookback == "" {
		t.Error("lookback param is empty")
	}
	if len(deps) != 2 {
		t.Errorf("got %d dependencies, want 2", len(deps))
	}
	if deps[0].Parent != "api-gateway" {
		t.Errorf("deps[0].Parent = %q, want %q", deps[0].Parent, "api-gateway")
	}
}

func TestGetTrace_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`trace not found`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.GetTrace("nonexistent", time.Time{}, time.Time{})
	if err == nil {
		t.Fatal("GetTrace() expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 error, got: %v", err)
	}
}

func TestSearchTraces_EmptyResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []interface{}{},
			"total":  0,
			"errors": []interface{}{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	results, err := c.SearchTraces("nonexistent-service", nil, "", time.Time{}, time.Time{}, 0, "", "", "")
	if err != nil {
		t.Fatalf("SearchTraces() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}
