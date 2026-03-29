package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// KratosSession holds the result of a successful Kratos login.
type KratosSession struct {
	Cookie    string    // full cookie header value to send on requests
	ExpiresAt time.Time // when the session expires
}

// KratosLogin performs the Ory Kratos browser login flow:
//  1. GET /api/auth/self-service/login/browser → flow ID + CSRF token
//  2. POST /api/auth/self-service/login?flow=<id> with credentials
//  3. Extract session cookie from the cookie jar
func KratosLogin(serverBaseURL, email, password string, verbose bool) (*KratosSession, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		// Do not follow redirects automatically for the initial flow request
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	baseURL := strings.TrimRight(serverBaseURL, "/")

	// Step 1: Initiate login flow
	flowURL := baseURL + "/api/auth/self-service/login/browser"
	if verbose {
		fmt.Printf("> GET %s\n", flowURL)
	}

	resp, err := client.Get(flowURL)
	if err != nil {
		return nil, fmt.Errorf("initiating login flow: %w", err)
	}
	defer resp.Body.Close()

	// The browser flow returns a 303 redirect; we need the flow ID from the Location header.
	// But we also need to fetch the flow details to get the CSRF token.
	var flowID string
	if resp.StatusCode == http.StatusSeeOther || resp.StatusCode == http.StatusFound {
		loc := resp.Header.Get("Location")
		u, err := url.Parse(loc)
		if err == nil {
			flowID = u.Query().Get("flow")
		}
	}

	if flowID == "" {
		// Try parsing the response body as JSON (API might return flow directly)
		body, _ := io.ReadAll(resp.Body)
		var flowResp struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(body, &flowResp) == nil && flowResp.ID != "" {
			flowID = flowResp.ID
		}
	}

	if flowID == "" {
		return nil, fmt.Errorf("could not extract login flow ID (HTTP %d)", resp.StatusCode)
	}

	if verbose {
		fmt.Printf("  Flow ID: %s\n", flowID)
	}

	// Step 1b: Fetch the flow details to get the CSRF token
	// Allow redirects for this request
	flowClient := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	flowDetailURL := baseURL + "/api/auth/self-service/login/flows?id=" + flowID
	if verbose {
		fmt.Printf("> GET %s\n", flowDetailURL)
	}

	resp2, err := flowClient.Get(flowDetailURL)
	if err != nil {
		return nil, fmt.Errorf("fetching login flow details: %w", err)
	}
	defer resp2.Body.Close()

	flowBody, err := io.ReadAll(resp2.Body)
	if err != nil {
		return nil, fmt.Errorf("reading flow response: %w", err)
	}

	csrfToken, err := extractCSRFToken(flowBody)
	if err != nil {
		return nil, fmt.Errorf("extracting CSRF token: %w", err)
	}

	// Step 2: Submit credentials
	loginURL := baseURL + "/api/auth/self-service/login?flow=" + flowID
	if verbose {
		fmt.Printf("> POST %s\n", loginURL)
	}

	formData := url.Values{
		"identifier": {email},
		"password":   {password},
		"csrf_token": {csrfToken},
		"method":     {"password"},
	}

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp3, err := flowClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("submitting login: %w", err)
	}
	defer resp3.Body.Close()

	loginBody, err := io.ReadAll(resp3.Body)
	if err != nil {
		return nil, fmt.Errorf("reading login response: %w", err)
	}

	if verbose {
		fmt.Printf("< %d %s\n", resp3.StatusCode, resp3.Status)
	}

	if resp3.StatusCode == http.StatusUnauthorized || resp3.StatusCode == http.StatusBadRequest {
		// Parse error from Kratos
		var errResp struct {
			UI struct {
				Messages []struct {
					Text string `json:"text"`
				} `json:"messages"`
			} `json:"ui"`
			Error struct {
				Message string `json:"message"`
				Reason  string `json:"reason"`
			} `json:"error"`
		}
		if json.Unmarshal(loginBody, &errResp) == nil {
			if len(errResp.UI.Messages) > 0 {
				return nil, fmt.Errorf("login failed: %s", errResp.UI.Messages[0].Text)
			}
			if errResp.Error.Message != "" {
				return nil, fmt.Errorf("login failed: %s", errResp.Error.Message)
			}
		}
		return nil, fmt.Errorf("login failed (HTTP %d): invalid email or password", resp3.StatusCode)
	}

	if resp3.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed (HTTP %d): %s", resp3.StatusCode, string(loginBody))
	}

	// Extract session expiry from response
	var sessionResp struct {
		Session struct {
			ExpiresAt string `json:"expires_at"`
		} `json:"session"`
	}
	expiresAt := time.Now().Add(24 * time.Hour) // default
	if json.Unmarshal(loginBody, &sessionResp) == nil && sessionResp.Session.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, sessionResp.Session.ExpiresAt); err == nil {
			expiresAt = t
		}
	}

	// Extract session cookies from the jar
	u, _ := url.Parse(baseURL)
	cookies := jar.Cookies(u)
	if len(cookies) == 0 {
		return nil, fmt.Errorf("login succeeded but no session cookies were set")
	}

	// Build cookie header string
	cookieParts := make([]string, 0, len(cookies))
	for _, c := range cookies {
		cookieParts = append(cookieParts, c.Name+"="+c.Value)
	}
	cookieHeader := strings.Join(cookieParts, "; ")

	return &KratosSession{
		Cookie:    cookieHeader,
		ExpiresAt: expiresAt,
	}, nil
}

// extractCSRFToken parses the Kratos flow JSON and finds the csrf_token value.
func extractCSRFToken(flowBody []byte) (string, error) {
	var flow struct {
		UI struct {
			Nodes []struct {
				Attributes struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"attributes"`
			} `json:"nodes"`
		} `json:"ui"`
	}

	if err := json.Unmarshal(flowBody, &flow); err != nil {
		return "", fmt.Errorf("parsing flow JSON: %w", err)
	}

	for _, node := range flow.UI.Nodes {
		if node.Attributes.Name == "csrf_token" {
			return node.Attributes.Value, nil
		}
	}

	return "", fmt.Errorf("csrf_token not found in login flow")
}
