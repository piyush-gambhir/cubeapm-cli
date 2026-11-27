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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/auth"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/config"
)

type Client struct {
	queryBaseURL  string
	ingestBaseURL string
	adminBaseURL  string
	serverBaseURL string
	httpClient    *http.Client
	verbose       bool

	authMethod    string
	sessionCookie string
	email         string
	password      string

	onSessionRefresh func(cookie string, expiry string) error
}

func (c *Client) SetOnSessionRefresh(fn func(cookie string, expiry string) error) {
	c.onSessionRefresh = fn
}

func NewClient(cfg config.ResolvedConfig) (*Client, error) {
	server := strings.TrimRight(cfg.Server, "/")
	if server == "" {
		return nil, fmt.Errorf("server address not configured; run 'cubeapm login' or set CUBEAPM_SERVER")
	}

	if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
		server = "http://" + server
	}

	u, err := url.Parse(server)
	if err != nil {
		return nil, fmt.Errorf("invalid server address %q: %w", server, err)
	}

	host := u.Hostname()
	scheme := u.Scheme

	serverBaseURL := (&url.URL{Scheme: scheme, Host: u.Host}).String()

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

	authMethod := cfg.AuthMethod
	if authMethod == "" {
		if cfg.Email != "" && cfg.Password != "" {
			authMethod = "kratos"
		} else {
			authMethod = "none"
		}
	}

	return &Client{
		queryBaseURL:  (&url.URL{Scheme: scheme, Host: net.JoinHostPort(host, strconv.Itoa(cfg.QueryPort))}).String(),
		ingestBaseURL: (&url.URL{Scheme: scheme, Host: net.JoinHostPort(host, strconv.Itoa(cfg.IngestPort))}).String(),
		adminBaseURL:  (&url.URL{Scheme: scheme, Host: net.JoinHostPort(host, strconv.Itoa(cfg.AdminPort))}).String(),
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

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	c.setAuthHeaders(req)

	if c.verbose {
		fmt.Fprintf(os.Stderr, "> %s %s\n", req.Method, req.URL.String())
		for name, values := range req.Header {
			for _, v := range values {
				fmt.Fprintf(os.Stderr, "> %s: %s\n", name, redactAuthHeaders(name, v))
			}
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "< %d %s\n", resp.StatusCode, resp.Status)
		for name, values := range resp.Header {
			for _, v := range values {
				fmt.Fprintf(os.Stderr, "< %s: %s\n", name, redactAuthHeaders(name, v))
			}
		}
	}

	if resp.StatusCode == http.StatusUnauthorized && c.authMethod == "kratos" && c.email != "" && c.password != "" {
		resp.Body.Close()

		if c.verbose {
			fmt.Fprintln(os.Stderr, "  Session expired, re-authenticating...")
		}

		if err := c.reAuthenticate(); err != nil {
			return nil, fmt.Errorf("re-authentication failed: %w", err)
		}

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
			fmt.Fprintf(os.Stderr, "< %d %s (after re-auth)\n", resp.StatusCode, resp.Status)
		}
	}

	return resp, nil
}

func (c *Client) setAuthHeaders(req *http.Request) {
	if c.authMethod == "kratos" && c.sessionCookie != "" {
		req.Header.Set("Cookie", c.sessionCookie)
	}
}

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

func cloneRequest(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.Body == nil || req.Body == http.NoBody {
		return clone, nil
	}
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		clone.Body = body
		return clone, nil
	}
	if seeker, ok := req.Body.(io.ReadSeeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		clone.Body = io.NopCloser(seeker)
		return clone, nil
	}
	return nil, fmt.Errorf("cannot retry %s %s: its body cannot be replayed", req.Method, req.URL.Path)
}

func redactAuthHeaders(headerName, value string) string {
	lower := strings.ToLower(headerName)
	switch lower {
	case "authorization", "cookie", "set-cookie":
		return "[REDACTED]"
	default:
		return value
	}
}

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

func (c *Client) postRaw(baseURL, path, contentType string, body io.Reader) (*http.Response, error) {
	u := baseURL + path

	var reqBody io.ReadSeeker
	if body != nil {
		data, err := readAllLimited(body)
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

func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, err := readAllLimited(resp.Body)
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

func (c *Client) streamJSON(resp *http.Response, handler func(json.RawMessage) error) error {
	scanner := bufio.NewScanner(resp.Body)
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

func (c *Client) QueryBaseURL() string {
	return c.queryBaseURL
}
