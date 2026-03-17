package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

// Client is the HTTP client for CubeAPM API.
type Client struct {
	queryBaseURL  string // http://server:3140
	ingestBaseURL string // http://server:3130
	adminBaseURL  string // http://server:3199
	token         string
	httpClient    *http.Client
	verbose       bool
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

	return &Client{
		queryBaseURL:  fmt.Sprintf("%s://%s:%d", scheme, host, cfg.QueryPort),
		ingestBaseURL: fmt.Sprintf("%s://%s:%d", scheme, host, cfg.IngestPort),
		adminBaseURL:  fmt.Sprintf("%s://%s:%d", scheme, host, cfg.AdminPort),
		token:         cfg.Token,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		verbose: cfg.Verbose,
	}, nil
}

// doRequest executes an HTTP request with auth and error handling.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	if c.verbose {
		fmt.Printf("> %s %s\n", req.Method, req.URL.String())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if c.verbose {
		fmt.Printf("< %d %s\n", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

// get performs a GET request and returns the response body.
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
func (c *Client) post(baseURL, path string, params url.Values) (*http.Response, error) {
	u := baseURL + path

	var body io.Reader
	if params != nil {
		body = strings.NewReader(params.Encode())
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
func (c *Client) postRaw(baseURL, path, contentType string, body io.Reader) (*http.Response, error) {
	u := baseURL + path

	req, err := http.NewRequest("POST", u, body)
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

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("resource not found (HTTP 404): the requested endpoint may not be available on this CubeAPM server. Response: %s", string(body))
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed (HTTP 401): check your token with 'cubeapm login'")
	case http.StatusForbidden:
		return fmt.Errorf("access denied (HTTP 403): your token may not have sufficient permissions")
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
