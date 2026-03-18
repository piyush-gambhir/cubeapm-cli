package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIngestMetrics_Prometheus(t *testing.T) {
	var gotPath, gotMethod, gotContentType string
	var gotBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data := strings.NewReader("http_requests_total{method=\"GET\"} 100\n")
	err := c.IngestMetrics("prometheus", data)
	if err != nil {
		t.Fatalf("IngestMetrics() error = %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/v1/import/prometheus" {
		t.Errorf("path = %q, want /api/v1/import/prometheus", gotPath)
	}
	if gotContentType != "text/plain" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "text/plain")
	}
	if !strings.Contains(gotBody, "http_requests_total") {
		t.Errorf("body does not contain expected metrics data")
	}
}

func TestIngestMetrics_OTLP(t *testing.T) {
	var gotPath, gotContentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data := strings.NewReader("protobuf-encoded-data")
	err := c.IngestMetrics("otlp", data)
	if err != nil {
		t.Fatalf("IngestMetrics() error = %v", err)
	}

	if gotPath != "/v1/metrics" {
		t.Errorf("path = %q, want /v1/metrics", gotPath)
	}
	if gotContentType != "application/x-protobuf" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "application/x-protobuf")
	}
}

func TestIngestLogs_JSONLine(t *testing.T) {
	var gotPath, gotContentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data := strings.NewReader(`{"_msg":"test log","level":"info"}` + "\n")
	err := c.IngestLogs("jsonline", data)
	if err != nil {
		t.Fatalf("IngestLogs() error = %v", err)
	}

	if gotPath != "/api/logs/insert/jsonline" {
		t.Errorf("path = %q, want /api/logs/insert/jsonline", gotPath)
	}
	if gotContentType != "application/stream+json" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "application/stream+json")
	}
}

func TestIngestLogs_OTLP(t *testing.T) {
	var gotPath, gotContentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data := strings.NewReader("protobuf-logs")
	err := c.IngestLogs("otlp", data)
	if err != nil {
		t.Fatalf("IngestLogs() error = %v", err)
	}

	if gotPath != "/v1/logs" {
		t.Errorf("path = %q, want /v1/logs", gotPath)
	}
	if gotContentType != "application/x-protobuf" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "application/x-protobuf")
	}
}

func TestIngestLogs_Loki(t *testing.T) {
	var gotPath, gotContentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data := strings.NewReader(`{"streams":[]}`)
	err := c.IngestLogs("loki", data)
	if err != nil {
		t.Fatalf("IngestLogs() error = %v", err)
	}

	if gotPath != "/api/logs/insert/loki/api/v1/push" {
		t.Errorf("path = %q, want /api/logs/insert/loki/api/v1/push", gotPath)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "application/json")
	}
}

func TestIngestLogs_Elastic(t *testing.T) {
	var gotPath, gotContentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	data := strings.NewReader(`{"index":{}}` + "\n" + `{"msg":"test"}` + "\n")
	err := c.IngestLogs("elastic", data)
	if err != nil {
		t.Fatalf("IngestLogs() error = %v", err)
	}

	if gotPath != "/api/logs/insert/elasticsearch/_bulk" {
		t.Errorf("path = %q, want /api/logs/insert/elasticsearch/_bulk", gotPath)
	}
	if gotContentType != "application/x-ndjson" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "application/x-ndjson")
	}
}

func TestIngestMetrics_UnsupportedFormat(t *testing.T) {
	c := &Client{
		ingestBaseURL: "http://localhost:3130",
		httpClient:    http.DefaultClient,
	}

	err := c.IngestMetrics("xml", strings.NewReader("data"))
	if err == nil {
		t.Fatal("IngestMetrics() expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported metrics format") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("error should mention the format: %v", err)
	}
}

func TestIngestLogs_UnsupportedFormat(t *testing.T) {
	c := &Client{
		ingestBaseURL: "http://localhost:3130",
		httpClient:    http.DefaultClient,
	}

	err := c.IngestLogs("xml", strings.NewReader("data"))
	if err == nil {
		t.Fatal("IngestLogs() expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported logs format") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("error should mention the format: %v", err)
	}
}
