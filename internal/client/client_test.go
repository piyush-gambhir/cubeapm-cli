package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := config.ResolvedConfig{
		Server:     "cubeapm.example.com",
		QueryPort:  3140,
		IngestPort: 3130,
		AdminPort:  3199,
		Token:      "test-token",
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c == nil {
		t.Fatal("NewClient() returned nil client")
	}
	if c.token != "test-token" {
		t.Errorf("token = %q, want %q", c.token, "test-token")
	}
}

func TestNewClient_EmptyServer(t *testing.T) {
	cfg := config.ResolvedConfig{
		Server:     "",
		QueryPort:  3140,
		IngestPort: 3130,
		AdminPort:  3199,
	}

	_, err := NewClient(cfg)
	if err == nil {
		t.Fatal("NewClient() expected error for empty server, got nil")
	}
	if !strings.Contains(err.Error(), "server address not configured") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestQueryURL(t *testing.T) {
	cfg := config.ResolvedConfig{
		Server:     "cubeapm.example.com",
		QueryPort:  3140,
		IngestPort: 3130,
		AdminPort:  3199,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	want := "http://cubeapm.example.com:3140"
	if c.queryBaseURL != want {
		t.Errorf("queryBaseURL = %q, want %q", c.queryBaseURL, want)
	}
}

func TestIngestURL(t *testing.T) {
	cfg := config.ResolvedConfig{
		Server:     "cubeapm.example.com",
		QueryPort:  3140,
		IngestPort: 3130,
		AdminPort:  3199,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	want := "http://cubeapm.example.com:3130"
	if c.ingestBaseURL != want {
		t.Errorf("ingestBaseURL = %q, want %q", c.ingestBaseURL, want)
	}
}

func TestAdminURL(t *testing.T) {
	cfg := config.ResolvedConfig{
		Server:     "cubeapm.example.com",
		QueryPort:  3140,
		IngestPort: 3130,
		AdminPort:  3199,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	want := "http://cubeapm.example.com:3199"
	if c.adminBaseURL != want {
		t.Errorf("adminBaseURL = %q, want %q", c.adminBaseURL, want)
	}
}

func TestBearerAuth_Header(t *testing.T) {
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := &Client{
		queryBaseURL:  ts.URL,
		ingestBaseURL: ts.URL,
		adminBaseURL:  ts.URL,
		token:         "my-secret-token",
		httpClient:    ts.Client(),
	}

	_, err := c.get(c.queryBaseURL, "/test", nil)
	if err != nil {
		t.Fatalf("get() error = %v", err)
	}

	want := "Bearer my-secret-token"
	if gotAuth != want {
		t.Errorf("Authorization header = %q, want %q", gotAuth, want)
	}
}

func TestErrorHandling_400(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`bad request body`))
	}))
	defer ts.Close()

	c := &Client{
		queryBaseURL: ts.URL,
		httpClient:   ts.Client(),
	}

	resp, err := c.get(c.queryBaseURL, "/test", nil)
	if err != nil {
		t.Fatalf("get() error = %v", err)
	}
	defer resp.Body.Close()

	err = c.checkResponse(resp)
	if err == nil {
		t.Fatal("checkResponse() expected error for 400 status")
	}
	if !strings.Contains(err.Error(), "API error (HTTP 400)") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "bad request body") {
		t.Errorf("expected error to contain response body, got: %v", err)
	}
}

func TestErrorHandling_401(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`unauthorized`))
	}))
	defer ts.Close()

	c := &Client{
		queryBaseURL: ts.URL,
		httpClient:   ts.Client(),
	}

	resp, err := c.get(c.queryBaseURL, "/test", nil)
	if err != nil {
		t.Fatalf("get() error = %v", err)
	}
	defer resp.Body.Close()

	err = c.checkResponse(resp)
	if err == nil {
		t.Fatal("checkResponse() expected error for 401 status")
	}
	if !strings.Contains(err.Error(), "authentication failed (HTTP 401)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestErrorHandling_500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal server error`))
	}))
	defer ts.Close()

	c := &Client{
		queryBaseURL: ts.URL,
		httpClient:   ts.Client(),
	}

	resp, err := c.get(c.queryBaseURL, "/test", nil)
	if err != nil {
		t.Fatalf("get() error = %v", err)
	}
	defer resp.Body.Close()

	err = c.checkResponse(resp)
	if err == nil {
		t.Fatal("checkResponse() expected error for 500 status")
	}
	if !strings.Contains(err.Error(), "API error (HTTP 500)") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "internal server error") {
		t.Errorf("expected error to contain response body, got: %v", err)
	}
}
