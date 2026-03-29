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
		Email:      "user@test.com",
		Password:   "secret",
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c == nil {
		t.Fatal("NewClient() returned nil client")
	}
	if c.authMethod != "kratos" {
		t.Errorf("authMethod = %q, want %q", c.authMethod, "kratos")
	}
	if c.email != "user@test.com" {
		t.Errorf("email = %q, want %q", c.email, "user@test.com")
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

func TestKratosAuth_CookieHeader(t *testing.T) {
	var gotCookie string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := &Client{
		queryBaseURL:  ts.URL,
		ingestBaseURL: ts.URL,
		adminBaseURL:  ts.URL,
		authMethod:    "kratos",
		sessionCookie: "ory_kratos_session=abc123",
		httpClient:    ts.Client(),
	}

	_, err := c.get(c.queryBaseURL, "/test", nil)
	if err != nil {
		t.Fatalf("get() error = %v", err)
	}

	want := "ory_kratos_session=abc123"
	if gotCookie != want {
		t.Errorf("Cookie header = %q, want %q", gotCookie, want)
	}
}

func TestNoAuth_NoHeaders(t *testing.T) {
	var gotAuth, gotCookie string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCookie = r.Header.Get("Cookie")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := &Client{
		queryBaseURL:  ts.URL,
		ingestBaseURL: ts.URL,
		adminBaseURL:  ts.URL,
		authMethod:    "none",
		httpClient:    ts.Client(),
	}

	_, err := c.get(c.queryBaseURL, "/test", nil)
	if err != nil {
		t.Fatalf("get() error = %v", err)
	}

	if gotAuth != "" {
		t.Errorf("Authorization header = %q, want empty", gotAuth)
	}
	if gotCookie != "" {
		t.Errorf("Cookie header = %q, want empty", gotCookie)
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

// --- Multi-port URL extraction tests ---

func TestNewClient_ServerWithPort_ExtractsHostname(t *testing.T) {
	cfg := config.ResolvedConfig{
		Server:     "https://my-server:9443",
		QueryPort:  3140,
		IngestPort: 3130,
		AdminPort:  3199,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	wantQuery := "https://my-server:3140"
	wantIngest := "https://my-server:3130"
	wantAdmin := "https://my-server:3199"

	if c.queryBaseURL != wantQuery {
		t.Errorf("queryBaseURL = %q, want %q", c.queryBaseURL, wantQuery)
	}
	if c.ingestBaseURL != wantIngest {
		t.Errorf("ingestBaseURL = %q, want %q", c.ingestBaseURL, wantIngest)
	}
	if c.adminBaseURL != wantAdmin {
		t.Errorf("adminBaseURL = %q, want %q", c.adminBaseURL, wantAdmin)
	}
}

func TestNewClient_ServerWithSchemeNoPort(t *testing.T) {
	cfg := config.ResolvedConfig{
		Server:     "https://secure.example.com",
		QueryPort:  3140,
		IngestPort: 3130,
		AdminPort:  3199,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if c.queryBaseURL != "https://secure.example.com:3140" {
		t.Errorf("queryBaseURL = %q, want %q", c.queryBaseURL, "https://secure.example.com:3140")
	}
}

func TestNewClient_ServerBareHostname(t *testing.T) {
	cfg := config.ResolvedConfig{
		Server:     "my-server",
		QueryPort:  3140,
		IngestPort: 3130,
		AdminPort:  3199,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if c.queryBaseURL != "http://my-server:3140" {
		t.Errorf("queryBaseURL = %q, want %q", c.queryBaseURL, "http://my-server:3140")
	}
}

func TestNewClient_ServerBareHostnameWithPort(t *testing.T) {
	cfg := config.ResolvedConfig{
		Server:     "my-server:3140",
		QueryPort:  3140,
		IngestPort: 3130,
		AdminPort:  3199,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Port in the server address should be stripped; configured ports are used.
	if c.queryBaseURL != "http://my-server:3140" {
		t.Errorf("queryBaseURL = %q, want %q", c.queryBaseURL, "http://my-server:3140")
	}
	if c.ingestBaseURL != "http://my-server:3130" {
		t.Errorf("ingestBaseURL = %q, want %q", c.ingestBaseURL, "http://my-server:3130")
	}
}

// --- Token redaction tests ---

func TestRedactAuthHeaders(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"Authorization", "Bearer secret-token", "[REDACTED]"},
		{"authorization", "Bearer secret-token", "[REDACTED]"},
		{"Cookie", "session=abc123", "[REDACTED]"},
		{"Set-Cookie", "session=abc123; Path=/", "[REDACTED]"},
		{"Content-Type", "application/json", "application/json"},
		{"X-Request-Id", "req-123", "req-123"},
	}

	for _, tt := range tests {
		got := redactAuthHeaders(tt.name, tt.value)
		if got != tt.want {
			t.Errorf("redactAuthHeaders(%q, %q) = %q, want %q", tt.name, tt.value, got, tt.want)
		}
	}
}

func TestVerboseOutput_RedactsAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := &Client{
		queryBaseURL:  ts.URL,
		ingestBaseURL: ts.URL,
		adminBaseURL:  ts.URL,
		authMethod:    "kratos",
		sessionCookie: "ory_kratos_session=super-secret",
		httpClient:    ts.Client(),
		verbose:       true,
	}

	// Just verify it doesn't panic with verbose mode on.
	_, err := c.get(c.queryBaseURL, "/test", nil)
	if err != nil {
		t.Fatalf("get() error = %v", err)
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
