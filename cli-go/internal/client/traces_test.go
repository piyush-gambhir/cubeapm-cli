package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/types"
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
	results, err := c.SearchTraces("api-gateway", nil, "", time.Time{}, time.Time{}, 0, "", "", "", "")
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

	_, err := c.SearchTraces("payments", tags, "POST /orders", from, to, 50, "server", "500ms", "2s", "")
	if err != nil {
		t.Fatalf("SearchTraces() error = %v", err)
	}

	if gotQuery.Get("service") != "payments" {
		t.Errorf("service = %q, want %q", gotQuery.Get("service"), "payments")
	}
	if gotQuery.Get("query") != "POST /orders" {
		t.Errorf("query = %q, want %q", gotQuery.Get("query"), "POST /orders")
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
	if gotQuery.Get("start") != "1705312800" {
		t.Errorf("start = %q, want Unix seconds 1705312800", gotQuery.Get("start"))
	}
	if gotQuery.Get("end") != "1705316400" {
		t.Errorf("end = %q, want Unix seconds 1705316400", gotQuery.Get("end"))
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

	_, err := c.SearchTraces("api-gateway", tags, "", time.Time{}, time.Time{}, 0, "", "", "", "")
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
	_, err := c.SearchTraces("api-gateway", nil, "", time.Time{}, time.Time{}, 0, "", "100ms", "5s", "")
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
	results, err := c.SearchTraces("nonexistent-service", nil, "", time.Time{}, time.Time{}, 0, "", "", "", "")
	if err != nil {
		t.Fatalf("SearchTraces() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestDecodeCubeID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "32-char hex trace ID passes through",
			in:   "4bf92f3577b34da6a3ce929d0e0e4736",
			want: "4bf92f3577b34da6a3ce929d0e0e4736",
		},
		{
			name: "16-char hex span ID passes through",
			in:   "00f067aa0ba902b7",
			want: "00f067aa0ba902b7",
		},
		{
			name: "upper-case hex is normalized to lower-case",
			in:   "4BF92F3577B34DA6A3CE929D0E0E4736",
			want: "4bf92f3577b34da6a3ce929d0e0e4736",
		},
		{
			name: "standard base64 trace ID decodes to hex",
			in:   "S/kvNXezTaajzpKdDg5HNg==", // 16 bytes
			want: "4bf92f3577b34da6a3ce929d0e0e4736",
		},
		{
			name: "raw URL-safe base64 decodes to hex",
			in:   "S_kvNXezTaajzpKdDg5HNg", // same 16 bytes, no padding
			want: "4bf92f3577b34da6a3ce929d0e0e4736",
		},
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "non-hex non-base64 returned as-is",
			in:   "not-an-id!",
			want: "not-an-id!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeCubeID(tt.in); got != tt.want {
				t.Errorf("decodeCubeID(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGetTrace_NativeShape(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CubeAPM native trace-get: {spans:[{...,process:{service_name}}], process_map:[]}
		w.Write([]byte(`{"spans":[{"trace_id":"e6319082328509fe1e0aa6734da74b2b","span_id":"1e0aa6734da74b2b","operation_name":"GET /x","start_time":"2026-06-14T11:09:32Z","duration":1234,"process":{"service_name":"CONSOLE-BACKEND"}}],"process_map":[]}`))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	tr, err := c.GetTrace("e6319082328509fe1e0aa6734da74b2b", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetTrace() error = %v", err)
	}
	if len(tr.Spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(tr.Spans))
	}
	if tr.Spans[0].OperationName != "GET /x" {
		t.Errorf("operation = %q, want %q", tr.Spans[0].OperationName, "GET /x")
	}
	pid := tr.Spans[0].ProcessID
	if pid == "" {
		t.Fatalf("span ProcessID empty; inline process not synthesized")
	}
	if tr.Processes[pid].ServiceName != "CONSOLE-BACKEND" {
		t.Errorf("service = %q, want CONSOLE-BACKEND", tr.Processes[pid].ServiceName)
	}
}

func TestGetServicesByEnv(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if len(r.URL.Query()["match[]"]) == 0 {
			t.Errorf("expected match[] param for %s", r.URL.Path)
		}
		data := []string{"A", "B"} // label "service"
		if strings.Contains(r.URL.Path, "/label/service.name/") {
			data = []string{"B", "C"}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": data})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	svcs, err := c.GetServicesByEnv("PROD", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetServicesByEnv() error = %v", err)
	}
	if calls != 4 { // 2 labels x 2 env matchers
		t.Errorf("got %d label-values calls, want 4", calls)
	}
	got := map[string]bool{}
	for _, s := range svcs {
		got[s] = true
	}
	for _, want := range []string{"A", "B", "C"} {
		if !got[want] {
			t.Errorf("missing service %q in %v", want, svcs)
		}
	}
	if len(svcs) != 3 {
		t.Errorf("got %d services, want 3 (deduped)", len(svcs))
	}
}

func TestConvertCubeTagsPreservesZeroValues(t *testing.T) {
	var tags []types.CubeSpanTag
	data := []byte(`[
		{"key":"retries","v_int":0},
		{"key":"sampled","v_bool":false},
		{"key":"ratio","v_float":0},
		{"key":"empty","v_str":""}
	]`)
	if err := json.Unmarshal(data, &tags); err != nil {
		t.Fatal(err)
	}
	got := convertCubeTags(tags)
	if len(got) != 4 {
		t.Fatalf("got %d tags, want 4", len(got))
	}
	wantTypes := []string{"int64", "bool", "float64", "string"}
	wantValues := []interface{}{int64(0), false, float64(0), ""}
	for i := range got {
		if got[i].Type != wantTypes[i] || got[i].Value != wantValues[i] {
			t.Errorf("tag %d = (%s, %#v), want (%s, %#v)", i, got[i].Type, got[i].Value, wantTypes[i], wantValues[i])
		}
	}
}

// The public CubeAPM trace API takes Unix seconds, independently of the
// microsecond timestamps used internally by the Jaeger-compatible renderer.
func TestTraceFetchDocumentedTimeRange(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/traces/api/v1/traces/abc123" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("start") != "1761666720" || r.URL.Query().Get("end") != "1761670320" {
			t.Errorf("expected documented Unix seconds, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"spans":[{"trace_id":"abc123","span_id":"123","operation_name":"GET /x","start_time":"2025-10-28T15:52:06Z","duration":1000,"process":{"service_name":"orders"}}]}`))
	}))
	defer ts.Close()
	_, err := newTestClient(ts).GetTrace("abc123", time.Unix(1761666720, 0), time.Unix(1761670320, 0))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchTracesPreservesEnvironmentFilters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("env") != "staging" {
			t.Errorf("missing environment: %s", r.URL.RawQuery)
		}
		if r.URL.Query().Get("query") != "*" {
			t.Errorf("missing native query: %s", r.URL.RawQuery)
		}
		if r.URL.Query().Has("operation") {
			t.Error("sent legacy operation parameter")
		}
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()
	tags := map[string]string{"environment": "staging", "http.method": "GET"}
	for range 2 {
		_, err := newTestClient(ts).SearchTraces("orders", tags, "*", time.Time{}, time.Time{}, 10, "", "", "", "")
		if err != nil {
			t.Fatal(err)
		}
	}
	if tags["environment"] != "staging" || len(tags) != 2 {
		t.Fatalf("mutated caller tags: %v", tags)
	}
}
