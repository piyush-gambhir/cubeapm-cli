package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/internal/auth"
	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

// Client is the HTTP client for CubeAPM API.
type Client struct {
	queryBaseURL  string // http://server:3140
	ingestBaseURL string // http://server:3130
	adminBaseURL  string // http://server:3199
	serverBaseURL string // original server URL for auth flows (e.g. https://cube.spyne.ai)
	httpClient    *http.Client
	verbose       bool

	// Kratos session auth
	authMethod    string // "kratos" or "none"
	sessionCookie string
	email         string
	password      string

	// Callback to persist updated session after re-auth
	onSessionRefresh func(cookie string, expiry string) error
}

// SetOnSessionRefresh sets a callback that is invoked when the client
// re-authenticates and obtains a new session cookie.
func (c *Client) SetOnSessionRefresh(fn func(cookie string, expiry string) error) {
	c.onSessionRefresh = fn
}

// NewClient creates a new Client from a resolved configuration.
func NewClient(cfg config.ResolvedConfig) (*Client, error) {
	server := strings.TrimRight(cfg.Server, "/")
	if server == "" {
		return nil, fmt.Errorf("server address not configured; run 'cubeapm login' or set CUBEAPM_SERVER")
	}

	// Add scheme if not present
	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "http://" + server
	}

	// Parse to extract host without port
	u, err := url.Parse(server)
	if err != nil {
		return nil, fmt.Errorf("invalid server address %q: %w", server, err)
	}

	host := u.Hostname()
	scheme := u.Scheme

	// serverBaseURL is the original server URL (for Kratos auth flows)
	serverBaseURL := fmt.Sprintf("%s://%s", scheme, host)
	if u.Port() != "" {
		serverBaseURL = fmt.Sprintf("%s://%s:%s", scheme, host, u.Port())
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
	}

	// Determine auth method
	authMethod := cfg.AuthMethod
	if authMethod == "" {
		if cfg.Email != "" && cfg.Password != "" {
			authMethod = "kratos"
		} else {
			authMethod = "none"
		}
	}

	return &Client{
		queryBaseURL:  fmt.Sprintf("%s://%s:%d", scheme, host, cfg.QueryPort),
		ingestBaseURL: fmt.Sprintf("%s://%s:%d", scheme, host, cfg.IngestPort),
		adminBaseURL:  fmt.Sprintf("%s://%s:%d", scheme, host, cfg.AdminPort),
		serverBaseURL: serverBaseURL,
		httpClient: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		},
		verbose:       cfg.Verbose,
		authMethod:    authMethod,
		sessionCookie: cfg.SessionCookie,
		email:         cfg.Email,
		password:      cfg.Password,
	}, nil
}

// doRequest executes an HTTP request with auth and error handling.
// For Kratos auth, it auto-re-authenticates on 401 and retries once.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	c.setAuthHeaders(req)

	if c.verbose {
		fmt.Printf("> %s %s\n", req.Method, req.URL.String())
		for name, values := range req.Header {
			for _, v := range values {
				fmt.Printf("> %s: %s\n", name, redactAuthHeaders(name, v))
			}
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if c.verbose {
		fmt.Printf("< %d %s\n", resp.StatusCode, resp.Status)
		for name, values := range resp.Header {
			for _, v := range values {
				fmt.Printf("< %s: %s\n", name, redactAuthHeaders(name, v))
			}
		}
	}

	// Auto re-auth on 401 for Kratos auth
	if resp.StatusCode == http.StatusUnauthorized && c.authMethod == "kratos" && c.email != "" && c.password != "" {
		resp.Body.Close()

		if c.verbose {
			fmt.Println("  Session expired, re-authenticating...")
		}

		if err := c.reAuthenticate(); err != nil {
			return nil, fmt.Errorf("re-authentication failed: %w", err)
		}

		// Rebuild the request for retry (body may have been consumed)
		retryReq, err := cloneRequest(req)
		if err != nil {
			return nil, fmt.Errorf("rebuilding request for retry: %w", err)
		}
		c.setAuthHeaders(retryReq)

		resp, err = c.httpClient.Do(retryReq)
		if err != nil {
			return nil, fmt.Errorf("request failed after re-auth: %w", err)
		}

		if c.verbose {
			fmt.Printf("< %d %s (after re-auth)\n", resp.StatusCode, resp.Status)
		}
	}

	return resp, nil
}

// setAuthHeaders applies the appropriate auth header/cookie to the request.
func (c *Client) setAuthHeaders(req *http.Request) {
	if c.authMethod == "kratos" && c.sessionCookie != "" {
		req.Header.Set("Cookie", c.sessionCookie)
	}
}

// reAuthenticate performs a Kratos login and updates the client's session cookie.
func (c *Client) reAuthenticate() error {
	session, err := auth.KratosLogin(c.serverBaseURL, c.email, c.password, c.verbose)
	if err != nil {
		return err
	}
	c.sessionCookie = session.Cookie
	if c.onSessionRefresh != nil {
		return c.onSessionRefresh(session.Cookie, session.ExpiresAt.Format(time.RFC3339))
	}
	return nil
}

// cloneRequest creates a copy of the request. If the original request had a body,
// it is read into a buffer and both the original and clone get a fresh reader.
func cloneRequest(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.Body != nil && req.Body != http.NoBody {
		if seeker, ok := req.Body.(io.ReadSeeker); ok {
			seeker.Seek(0, io.SeekStart)
			clone.Body = io.NopCloser(seeker)
		}
		// If body was already consumed and not seekable, it will be empty.
		// This is acceptable for the retry case since GET requests have no body,
		// and POST requests in this CLI use bytes.NewReader which is seekable.
	}
	return clone, nil
}

// redactAuthHeaders returns a redacted value for sensitive headers
// (Authorization, Cookie, Set-Cookie) and the original value for all others.
func redactAuthHeaders(headerName, value string) string {
	lower := strings.ToLower(headerName)
	switch lower {
	case "authorization", "cookie", "set-cookie":
		return "[REDACTED]"
	default:
		return value
	}
}

// get performs a GET request and returns the response.
// The caller is responsible for closing the response body.
func (c *Client) get(baseURL, path string, params url.Values) (*http.Response, error) {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	return c.doRequest(req)
}

// post performs a POST request with form-encoded body and returns the response.
// The caller is responsible for closing the response body.
func (c *Client) post(baseURL, path string, params url.Values) (*http.Response, error) {
	u := baseURL + path

	var body io.ReadSeeker
	if params != nil {
		body = bytes.NewReader([]byte(params.Encode()))
	}

	req, err := http.NewRequest("POST", u, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if params != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return c.doRequest(req)
}

// postRaw performs a POST request with a raw body.
// The caller is responsible for closing the response body.
func (c *Client) postRaw(baseURL, path, contentType string, body io.Reader) (*http.Response, error) {
	u := baseURL + path

	// Buffer the body so it can be replayed on 401 retry
	var reqBody io.ReadSeeker
	if body != nil {
		data, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("buffering request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest("POST", u, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return c.doRequest(req)
}

// getJSON performs a GET request and decodes the JSON response into target.
func (c *Client) getJSON(baseURL, path string, params url.Values, target interface{}) error {
	resp, err := c.get(baseURL, path, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return err
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// postJSON performs a POST request and decodes the JSON response into target.
func (c *Client) postJSON(baseURL, path string, params url.Values, target interface{}) error {
	resp, err := c.post(baseURL, path, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return err
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// checkResponse checks the HTTP response for errors.
func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte(fmt.Sprintf("<failed to read response body: %v>", err))
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("resource not found (HTTP 404): the requested endpoint may not be available on this CubeAPM server. Response: %s", string(body))
	case http.StatusUnauthorized:
		if c.authMethod == "kratos" {
			return fmt.Errorf("authentication failed (HTTP 401): session expired or credentials invalid. Run 'cubeapm login' to re-authenticate")
		}
		return fmt.Errorf("authentication failed (HTTP 401): this CubeAPM instance requires authentication. Run 'cubeapm login'")
	case http.StatusForbidden:
		return fmt.Errorf("access denied (HTTP 403): insufficient permissions")
	default:
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}
}

// streamJSON reads newline-delimited JSON from the response and calls handler for each line.
func (c *Client) streamJSON(resp *http.Response, handler func(json.RawMessage) error) error {
	scanner := bufio.NewScanner(resp.Body)
	// Increase max line size for large log entries
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		if err := handler(json.RawMessage(line)); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// QueryBaseURL returns the query base URL (for testing connectivity).
func (c *Client) QueryBaseURL() string {
	return c.queryBaseURL
}
