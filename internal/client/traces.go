package client

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

const tracesBasePath = "/api/traces/api/v1"

// SearchTraces searches for traces matching the given criteria.
func (c *Client) SearchTraces(service, env, query string, from, to time.Time, limit int, spanKind string) ([]types.TraceSearchResult, error) {
	params := url.Values{}
	if service != "" {
		params.Set("service", service)
	}
	if env != "" {
		params.Set("tags", fmt.Sprintf(`{"environment":"%s"}`, env))
	}
	if query != "" {
		params.Set("operation", query)
	}
	if !from.IsZero() {
		params.Set("start", strconv.FormatInt(from.UnixMicro(), 10))
	}
	if !to.IsZero() {
		params.Set("end", strconv.FormatInt(to.UnixMicro(), 10))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if spanKind != "" {
		params.Set("spanKind", spanKind)
	}

	var resp types.JaegerSearchResponse
	if err := c.getJSON(c.queryBaseURL, tracesBasePath+"/search", params, &resp); err != nil {
		return nil, fmt.Errorf("searching traces: %w", err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("trace search error: %s", resp.Errors[0].Msg)
	}

	return resp.Data, nil
}

// GetTrace retrieves a trace by its trace ID.
func (c *Client) GetTrace(traceID string, from, to time.Time) (*types.Trace, error) {
	params := url.Values{}
	if !from.IsZero() {
		params.Set("start", strconv.FormatInt(from.UnixMicro(), 10))
	}
	if !to.IsZero() {
		params.Set("end", strconv.FormatInt(to.UnixMicro(), 10))
	}

	var resp types.JaegerTraceResponse
	path := fmt.Sprintf("%s/traces/%s", tracesBasePath, traceID)
	if err := c.getJSON(c.queryBaseURL, path, params, &resp); err != nil {
		return nil, fmt.Errorf("getting trace %s: %w", traceID, err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("trace error: %s", resp.Errors[0].Msg)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("trace %s not found", traceID)
	}

	return &resp.Data[0], nil
}

// GetServices returns a list of all services.
func (c *Client) GetServices() ([]string, error) {
	var resp types.JaegerServicesResponse
	if err := c.getJSON(c.queryBaseURL, tracesBasePath+"/services", nil, &resp); err != nil {
		return nil, fmt.Errorf("getting services: %w", err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("services error: %s", resp.Errors[0].Msg)
	}

	return resp.Data, nil
}

// GetOperations returns operations for a given service.
func (c *Client) GetOperations(service, spanKind string) ([]types.Operation, error) {
	params := url.Values{}
	if spanKind != "" {
		params.Set("spanKind", spanKind)
	}

	var resp types.JaegerOperationsResponse
	path := fmt.Sprintf("%s/services/%s/operations", tracesBasePath, url.PathEscape(service))
	if err := c.getJSON(c.queryBaseURL, path, params, &resp); err != nil {
		return nil, fmt.Errorf("getting operations for %s: %w", service, err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("operations error: %s", resp.Errors[0].Msg)
	}

	return resp.Data, nil
}

// GetDependencies returns service dependencies.
func (c *Client) GetDependencies(from, to time.Time) ([]types.Dependency, error) {
	params := url.Values{}
	if !from.IsZero() {
		params.Set("endTs", strconv.FormatInt(to.UnixMilli(), 10))
		lookback := to.Sub(from).Milliseconds()
		params.Set("lookback", strconv.FormatInt(lookback, 10))
	}

	var resp types.JaegerDependenciesResponse
	if err := c.getJSON(c.queryBaseURL, tracesBasePath+"/dependencies", params, &resp); err != nil {
		return nil, fmt.Errorf("getting dependencies: %w", err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("dependencies error: %s", resp.Errors[0].Msg)
	}

	return resp.Data, nil
}
