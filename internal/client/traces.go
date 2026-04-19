package client

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

const tracesBasePath = "/api/traces/api/v1"

// SearchTraces searches for traces matching the given criteria.
// The index parameter selects the CubeAPM backing store (e.g. "cube:latency",
// "cube:error"). It is required by the server; callers should pass a sensible
// default when users don't specify one.
func (c *Client) SearchTraces(service string, tags map[string]string, query string, from, to time.Time, limit int, spanKind, minDuration, maxDuration, index string) ([]types.TraceSearchResult, error) {
	params := url.Values{}
	if index != "" {
		params.Set("index", index)
	}
	// CubeAPM requires an "env" query param alongside tag-based filtering.
	// Promote tags["environment"] or tags["env"] to the top-level param so
	// --env on the CLI (or an explicit tag) satisfies the requirement.
	if env, ok := tags["environment"]; ok && env != "" {
		params.Set("env", env)
		delete(tags, "environment")
	} else if env, ok := tags["env"]; ok && env != "" {
		params.Set("env", env)
		delete(tags, "env")
	}
	if service != "" {
		params.Set("service", service)
	}
	if len(tags) > 0 {
		tagsJSON, err := json.Marshal(tags)
		if err != nil {
			return nil, fmt.Errorf("encoding tags: %w", err)
		}
		params.Set("tags", string(tagsJSON))
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
	if minDuration != "" {
		params.Set("minDuration", minDuration)
	}
	if maxDuration != "" {
		params.Set("maxDuration", maxDuration)
	}

	// Some CubeAPM servers return a raw JSON array (native format) while
	// classic Jaeger-compatible servers return {"data": [...]}. Read the
	// body once, then try both shapes before giving up. This lets the CLI
	// work against both deployments without a flag or probe request.
	body, err := c.getRaw(c.queryBaseURL, tracesBasePath+"/search", params)
	if err != nil {
		return nil, fmt.Errorf("searching traces: %w", err)
	}

	// Dispatch on the top-level JSON shape: classic Jaeger servers return
	// an object {data, errors, ...}; CubeAPM native returns an array.
	trimmed := firstNonSpace(body)
	if trimmed == '{' {
		var jaeger types.JaegerSearchResponse
		if err := json.Unmarshal(body, &jaeger); err != nil {
			return nil, fmt.Errorf("searching traces: %w", err)
		}
		if len(jaeger.Errors) > 0 {
			return nil, fmt.Errorf("trace search error: %s", jaeger.Errors[0].Msg)
		}
		return jaeger.Data, nil
	}

	var native types.CubeSearchResultList
	if err := json.Unmarshal(body, &native); err != nil {
		return nil, fmt.Errorf("searching traces: unable to decode response: %w", err)
	}
	return cubeToSearchResults(native), nil
}

// firstNonSpace returns the first non-whitespace byte of a JSON document
// so we can branch between object and array shapes without full parsing.
func firstNonSpace(b []byte) byte {
	for _, c := range b {
		switch c {
		case ' ', '\t', '\n', '\r':
			continue
		default:
			return c
		}
	}
	return 0
}

// cubeToSearchResults converts CubeAPM's native search shape into the
// Jaeger-style TraceSearchResult the CLI consumes elsewhere. We do this
// in the client so the command layer doesn't need to know about the
// server-specific format.
func cubeToSearchResults(in types.CubeSearchResultList) []types.TraceSearchResult {
	out := make([]types.TraceSearchResult, 0, len(in))
	for _, item := range in {
		spans := make([]types.Span, 0, len(item.Trace.Spans))
		var traceID string
		for _, cs := range item.Trace.Spans {
			tid := decodeCubeID(cs.TraceID)
			sid := decodeCubeID(cs.SpanID)
			if traceID == "" {
				traceID = tid
			}
			spans = append(spans, types.Span{
				TraceID:       tid,
				SpanID:        sid,
				OperationName: cs.OperationName,
				References:    convertCubeRefs(cs.References),
				StartTime:     parseCubeStart(cs.StartTime),
				Duration:      cs.Duration,
				Tags:          convertCubeTags(cs.Tags),
				ProcessID:     cs.ProcessID,
			})
		}
		out = append(out, types.TraceSearchResult{
			TraceID:   traceID,
			Spans:     spans,
			Processes: item.Trace.Processes,
		})
	}
	return out
}

// decodeCubeID accepts a trace or span ID in either raw hex, base64, or
// URL-safe base64 and returns the lower-case hex representation that
// Jaeger-style tooling expects. If the input doesn't decode cleanly as
// base64, we assume it was already a hex string.
func decodeCubeID(s string) string {
	if s == "" {
		return ""
	}
	// Try standard base64 then URL-safe base64 with padding tolerance.
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.RawURLEncoding} {
		if b, err := enc.DecodeString(s); err == nil && len(b) > 0 {
			return hex.EncodeToString(b)
		}
	}
	return s
}

// parseCubeStart turns CubeAPM's RFC3339-string start_time into the
// microseconds-since-epoch int64 that Jaeger consumers expect.
func parseCubeStart(s string) int64 {
	if s == "" {
		return 0
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UnixMicro()
	}
	return 0
}

func convertCubeRefs(in []types.CubeSpanRef) []types.SpanRef {
	if len(in) == 0 {
		return nil
	}
	out := make([]types.SpanRef, 0, len(in))
	for _, r := range in {
		out = append(out, types.SpanRef{
			RefType: r.RefType,
			TraceID: decodeCubeID(r.TraceID),
			SpanID:  decodeCubeID(r.SpanID),
		})
	}
	return out
}

// convertCubeTags flattens CubeAPM's typed tag format (v_str/v_int/...)
// into the generic key/value shape used by the rest of the CLI.
func convertCubeTags(in []types.CubeSpanTag) []types.KeyValue {
	if len(in) == 0 {
		return nil
	}
	out := make([]types.KeyValue, 0, len(in))
	for _, t := range in {
		kv := types.KeyValue{Key: t.Key}
		switch {
		case t.VStr != "":
			kv.Type = "string"
			kv.Value = t.VStr
		case t.VInt != 0:
			kv.Type = "int64"
			kv.Value = t.VInt
		case t.VBool:
			kv.Type = "bool"
			kv.Value = t.VBool
		case t.VFloat != 0:
			kv.Type = "float64"
			kv.Value = t.VFloat
		default:
			kv.Type = "string"
			kv.Value = ""
		}
		out = append(out, kv)
	}
	return out
}

// getRaw performs the same GET that getJSON does but returns the raw body
// so the caller can try multiple JSON shapes.
func (c *Client) getRaw(baseURL, path string, params url.Values) ([]byte, error) {
	resp, err := c.get(baseURL, path, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
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
//
// Some CubeAPM deployments don't expose the Jaeger /services endpoint (the
// reverse proxy only forwards the subset of paths the UI uses). When the
// Jaeger call fails, fall back to the Prometheus-standard metrics API:
// `label-values service` gives the same information and is a contract
// every CubeAPM metrics backend supports.
func (c *Client) GetServices() ([]string, error) {
	var resp types.JaegerServicesResponse
	jaegerErr := c.getJSON(c.queryBaseURL, tracesBasePath+"/services", nil, &resp)
	if jaegerErr == nil {
		if len(resp.Errors) > 0 {
			// Fall through to metrics fallback — we still have a usable alternative.
		} else {
			return resp.Data, nil
		}
	}

	// Fallback: derive the service list from metrics. Use the standard
	// Prometheus label-values endpoint.
	services, fbErr := c.GetLabelValues("service", time.Time{}, time.Time{})
	if fbErr == nil && len(services) > 0 {
		return services, nil
	}

	// Both paths failed — surface the original Jaeger error since it's
	// what the caller was actually asking for.
	if jaegerErr != nil {
		return nil, fmt.Errorf("getting services: %w", jaegerErr)
	}
	return nil, fmt.Errorf("services error: %s", resp.Errors[0].Msg)
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
