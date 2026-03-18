package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestInstantQuery(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		r.ParseForm()
		gotQuery = r.FormValue("query")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "vector",
				"result":     []interface{}{},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	result, err := c.QueryInstant("up", time.Time{})
	if err != nil {
		t.Fatalf("QueryInstant() error = %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/metrics/api/v1/query" {
		t.Errorf("path = %q, want /api/metrics/api/v1/query", gotPath)
	}
	if gotQuery != "up" {
		t.Errorf("query = %q, want %q", gotQuery, "up")
	}
	if result.Status != "success" {
		t.Errorf("status = %q, want %q", result.Status, "success")
	}
}

func TestInstantQuery_ResponseParsing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "vector",
				"result": []interface{}{
					map[string]interface{}{
						"metric": map[string]string{"__name__": "up", "job": "prometheus"},
						"value":  []interface{}{1705312800.0, "1"},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	result, err := c.QueryInstant("up", time.Time{})
	if err != nil {
		t.Fatalf("QueryInstant() error = %v", err)
	}

	if result.Status != "success" {
		t.Errorf("status = %q, want %q", result.Status, "success")
	}
	if result.Data.ResultType != "vector" {
		t.Errorf("resultType = %q, want %q", result.Data.ResultType, "vector")
	}
	if result.Data.Result == nil {
		t.Error("result is nil, expected vector data")
	}
}

func TestRangeQuery(t *testing.T) {
	var gotPath, gotMethod string
	var gotQuery, gotStep string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		r.ParseForm()
		gotQuery = r.FormValue("query")
		gotStep = r.FormValue("step")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "matrix",
				"result":     []interface{}{},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	result, err := c.QueryRange("rate(http_requests_total[5m])", from, to, "60s")
	if err != nil {
		t.Fatalf("QueryRange() error = %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/metrics/api/v1/query_range" {
		t.Errorf("path = %q, want /api/metrics/api/v1/query_range", gotPath)
	}
	if gotQuery != "rate(http_requests_total[5m])" {
		t.Errorf("query = %q, want %q", gotQuery, "rate(http_requests_total[5m])")
	}
	if gotStep != "60s" {
		t.Errorf("step = %q, want %q", gotStep, "60s")
	}
	if result.Data.ResultType != "matrix" {
		t.Errorf("resultType = %q, want %q", result.Data.ResultType, "matrix")
	}
}

func TestRangeQuery_ResponseParsing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "matrix",
				"result": []interface{}{
					map[string]interface{}{
						"metric": map[string]string{"__name__": "http_requests_total"},
						"values": []interface{}{
							[]interface{}{1705312800.0, "10"},
							[]interface{}{1705312860.0, "15"},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	result, err := c.QueryRange("http_requests_total", from, to, "60s")
	if err != nil {
		t.Fatalf("QueryRange() error = %v", err)
	}

	if result.Status != "success" {
		t.Errorf("status = %q, want %q", result.Status, "success")
	}
	if result.Data.ResultType != "matrix" {
		t.Errorf("resultType = %q, want %q", result.Data.ResultType, "matrix")
	}
}

func TestGetLabels(t *testing.T) {
	var gotPath, gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []string{"__name__", "job", "instance"},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	labels, err := c.GetLabels(time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetLabels() error = %v", err)
	}

	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/metrics/api/v1/labels" {
		t.Errorf("path = %q, want /api/metrics/api/v1/labels", gotPath)
	}
	if len(labels) != 3 {
		t.Errorf("got %d labels, want 3", len(labels))
	}
	if labels[0] != "__name__" {
		t.Errorf("labels[0] = %q, want %q", labels[0], "__name__")
	}
}

func TestGetLabelValues(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []string{"prometheus", "node-exporter", "grafana"},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	values, err := c.GetLabelValues("job", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetLabelValues() error = %v", err)
	}

	if gotPath != "/api/metrics/api/v1/label/job/values" {
		t.Errorf("path = %q, want /api/metrics/api/v1/label/job/values", gotPath)
	}
	if len(values) != 3 {
		t.Errorf("got %d values, want 3", len(values))
	}
}

func TestGetSeries(t *testing.T) {
	var gotPath, gotMethod string
	var gotMatch []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotMatch = r.URL.Query()["match[]"]
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": []map[string]string{
				{"__name__": "http_requests_total", "job": "api"},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	series, err := c.GetSeries([]string{"http_requests_total"}, time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Fatalf("GetSeries() error = %v", err)
	}

	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/metrics/api/v1/series" {
		t.Errorf("path = %q, want /api/metrics/api/v1/series", gotPath)
	}
	if len(gotMatch) != 1 || gotMatch[0] != "http_requests_total" {
		t.Errorf("match[] = %v, want [http_requests_total]", gotMatch)
	}
	if len(series) != 1 {
		t.Errorf("got %d series, want 1", len(series))
	}
}

func TestGetSeries_WithLimit(t *testing.T) {
	var gotLimit string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []map[string]string{},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.GetSeries([]string{"up"}, time.Time{}, time.Time{}, 50)
	if err != nil {
		t.Fatalf("GetSeries() error = %v", err)
	}

	if gotLimit != "50" {
		t.Errorf("limit = %q, want %q", gotLimit, "50")
	}
}

func TestInstantQuery_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "error",
			"errorType": "bad_data",
			"error":     "invalid PromQL expression",
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.QueryInstant("invalid{{{", time.Time{})
	if err == nil {
		t.Fatal("QueryInstant() expected error for error response, got nil")
	}
	if !strings.Contains(err.Error(), "invalid PromQL expression") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRangeQuery_EmptyResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"resultType": "matrix",
				"result":     []interface{}{},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient(ts)
	from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	result, err := c.QueryRange("nonexistent_metric", from, to, "60s")
	if err != nil {
		t.Fatalf("QueryRange() error = %v", err)
	}
	if result.Status != "success" {
		t.Errorf("status = %q, want %q", result.Status, "success")
	}
}
