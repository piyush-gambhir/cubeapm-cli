package client

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

// isUnsupportedPath reports whether an API error is CubeAPM's "unsupported
// path requested" 400. Some deployments only proxy the subset of the Jaeger
// query API the UI uses, so endpoints like operations/dependencies 400 there.
func isUnsupportedPath(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unsupported path")
}

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
		if strings.Contains(err.Error(), "env is required") {
			return nil, fmt.Errorf("this CubeAPM server requires an environment for trace search; pass --env (e.g. --env PROD)")
		}
		if strings.Contains(err.Error(), "service is required") {
			return nil, fmt.Errorf("this CubeAPM server requires a service for trace search; pass --service <name>")
		}
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
		spans, procs, traceID := buildSpansAndProcesses(item.Trace.Spans, item.Trace.Processes)
		out = append(out, types.TraceSearchResult{
			TraceID:   traceID,
			Spans:     spans,
			Processes: procs,
		})
	}
	return out
}

// buildSpansAndProcesses converts CubeAPM native spans into Jaeger-style spans
// plus a processes map, returning the first span's trace ID. CubeAPM commonly
// attaches the process INLINE on each span (process_map/process_id empty), so
// when a span carries an inline process we synthesize a process entry keyed by
// service name and point the span's ProcessID at it, restoring the SERVICE
// column the rest of the CLI resolves through the processes map.
func buildSpansAndProcesses(cubeSpans []types.CubeSpan, processes map[string]types.Process) ([]types.Span, map[string]types.Process, string) {
	procs := processes
	if procs == nil {
		procs = map[string]types.Process{}
	}
	spans := make([]types.Span, 0, len(cubeSpans))
	traceID := ""
	for _, cs := range cubeSpans {
		tid := decodeCubeID(cs.TraceID)
		if traceID == "" {
			traceID = tid
		}
		pid := cs.ProcessID
		if cs.Process != nil {
			sn := cs.Process.ServiceName
			if sn == "" {
				sn = cs.Process.ServiceNameSnake
			}
			if sn != "" {
				if pid == "" {
					pid = sn
				}
				if _, ok := procs[pid]; !ok {
					procs[pid] = types.Process{ServiceName: sn, Tags: convertCubeTags(cs.Process.Tags)}
				}
			}
		}
		spans = append(spans, types.Span{
			TraceID:       tid,
			SpanID:        decodeCubeID(cs.SpanID),
			OperationName: cs.OperationName,
			References:    convertCubeRefs(cs.References),
			StartTime:     parseCubeStart(cs.StartTime),
			Duration:      cs.Duration,
			Tags:          convertCubeTags(cs.Tags),
			ProcessID:     pid,
		})
	}
	return spans, procs, traceID
}

// decodeCubeID accepts a trace or span ID in either raw hex, base64, or
// URL-safe base64 and returns the lower-case hex representation that
// Jaeger-style tooling expects.
//
// Hex must be checked first: every 32-char hex trace ID (and 16-char hex
// span ID) is *also* valid base64 (length divisible by 4, all chars in the
// base64 alphabet), so trying base64 first silently corrupted hex IDs into
// garbage. A genuine base64 ID being pure hex characters is astronomically
// unlikely, so the hex-first order is safe.
func decodeCubeID(s string) string {
	if s == "" {
		return ""
	}
	if len(s)%2 == 0 {
		if b, err := hex.DecodeString(s); err == nil && len(b) > 0 {
			return hex.EncodeToString(b) // normalizes to lower-case
		}
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
//
// Classic Jaeger query servers return {"data":[Trace],"errors":[]} with hex
// IDs and microsecond start times. CubeAPM returns its NATIVE shape instead:
// {"spans":[...],"process_map":[...]} with snake_case fields, base64 IDs and
// RFC3339 start times, the same wire format the search endpoint uses. We read
// the raw body and handle both so `traces get` works on either deployment.
func (c *Client) GetTrace(traceID string, from, to time.Time) (*types.Trace, error) {
	params := url.Values{}
	if !from.IsZero() {
		params.Set("start", strconv.FormatInt(from.UnixMicro(), 10))
	}
	if !to.IsZero() {
		params.Set("end", strconv.FormatInt(to.UnixMicro(), 10))
	}

	path := fmt.Sprintf("%s/traces/%s", tracesBasePath, traceID)
	body, err := c.getRaw(c.queryBaseURL, path, params)
	if err != nil {
		return nil, fmt.Errorf("getting trace %s: %w", traceID, err)
	}

	var dual struct {
		Data   []types.Trace `json:"data"`
		Errors []struct {
			Msg string `json:"msg"`
		} `json:"errors"`
		Spans      []types.CubeSpan         `json:"spans"`
		ProcessMap json.RawMessage          `json:"process_map"`
		Processes  map[string]types.Process `json:"processes"`
	}
	if err := json.Unmarshal(body, &dual); err != nil {
		return nil, fmt.Errorf("getting trace %s: %w", traceID, err)
	}
	if len(dual.Errors) > 0 {
		return nil, fmt.Errorf("trace error: %s", dual.Errors[0].Msg)
	}
	// Classic Jaeger shape.
	if len(dual.Data) > 0 {
		return &dual.Data[0], nil
	}
	// CubeAPM native shape.
	if len(dual.Spans) > 0 {
		return cubeSpansToTrace(traceID, dual.Spans, dual.ProcessMap, dual.Processes), nil
	}
	return nil, fmt.Errorf("trace %s not found", traceID)
}

// cubeSpansToTrace converts CubeAPM's native trace-get payload into the
// Jaeger-style types.Trace the CLI renders, reusing the same span/tag/id
// converters as the search path.
func cubeSpansToTrace(traceID string, cubeSpans []types.CubeSpan, processMapRaw json.RawMessage, processes map[string]types.Process) *types.Trace {
	procs := processes
	if len(procs) == 0 {
		procs = parseCubeProcessMap(processMapRaw)
	}
	spans, procs, tid := buildSpansAndProcesses(cubeSpans, procs)
	if traceID == "" {
		traceID = tid
	}
	return &types.Trace{TraceID: traceID, Spans: spans, Processes: procs}
}

// parseCubeProcessMap leniently decodes CubeAPM's process_map, which has been
// seen as a JSON array of {process_id|key, process|value:{service_name|serviceName, tags}}
// and (on classic servers) as an object {pID: {serviceName, tags}}. Best-effort:
// spans still render even if the process map can't be resolved.
func parseCubeProcessMap(raw json.RawMessage) map[string]types.Process {
	if len(raw) == 0 {
		return nil
	}
	type proc struct {
		ServiceNameSnake string              `json:"service_name"`
		ServiceName      string              `json:"serviceName"`
		Tags             []types.CubeSpanTag `json:"tags"`
	}
	pick := func(p proc) types.Process {
		sn := p.ServiceName
		if sn == "" {
			sn = p.ServiceNameSnake
		}
		return types.Process{ServiceName: sn, Tags: convertCubeTags(p.Tags)}
	}
	out := map[string]types.Process{}
	// Array form.
	var arr []struct {
		ProcessID string `json:"process_id"`
		Key       string `json:"key"`
		Process   *proc  `json:"process"`
		Value     *proc  `json:"value"`
	}
	if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
		for _, e := range arr {
			id := e.ProcessID
			if id == "" {
				id = e.Key
			}
			p := e.Process
			if p == nil {
				p = e.Value
			}
			if id == "" || p == nil {
				continue
			}
			out[id] = pick(*p)
		}
		if len(out) > 0 {
			return out
		}
	}
	// Object form.
	var obj map[string]proc
	if err := json.Unmarshal(raw, &obj); err == nil {
		for id, p := range obj {
			out[id] = pick(p)
		}
	}
	return out
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
			// Fall through to metrics fallback, we still have a usable alternative.
		} else {
			return resp.Data, nil
		}
	}

	// Fallback: derive the service list from metrics. Use the standard
	// Prometheus label-values endpoint.
	services, fbErr := c.GetLabelValues("service", nil, time.Time{}, time.Time{})
	if fbErr == nil && len(services) > 0 {
		return services, nil
	}

	// Both paths failed, surface the original Jaeger error since it's
	// what the caller was actually asking for.
	if jaegerErr != nil {
		return nil, fmt.Errorf("getting services: %w", jaegerErr)
	}
	return nil, fmt.Errorf("services error: %s", resp.Errors[0].Msg)
}

// GetServicesByEnv returns the distinct service names seen in a given
// environment. CubeAPM labels the environment as `env` on some metrics and
// `cube.environment` on others, and the service appears under both `service`
// and `service.name`, so we union label-values across every combination via
// the metrics label-values match[] extension.
//
// Caveat: this is metrics-derived, so a service that emits only traces (no
// metrics) in that environment may not appear. Each matcher is queried
// independently so one unsupported label/selector degrades gracefully.
func (c *Client) GetServicesByEnv(env string, from, to time.Time) ([]string, error) {
	matchers := []string{
		fmt.Sprintf(`{env=%q}`, env),
		fmt.Sprintf(`{cube.environment=%q}`, env),
	}
	seen := map[string]bool{}
	var out []string
	var lastErr error
	for _, label := range []string{"service", "service.name"} {
		for _, m := range matchers {
			vals, err := c.GetLabelValues(label, []string{m}, from, to)
			if err != nil {
				lastErr = err
				continue
			}
			for _, v := range vals {
				if v != "" && !seen[v] {
					seen[v] = true
					out = append(out, v)
				}
			}
		}
	}
	if len(out) == 0 && lastErr != nil {
		return nil, fmt.Errorf("getting services for env %q: %w", env, lastErr)
	}
	return out, nil
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
		if isUnsupportedPath(err) {
			return nil, fmt.Errorf("the operations endpoint is not exposed by this CubeAPM deployment; run 'cubeapm traces search --service %s --env <ENV>' and read the OPERATION column to see recent operations", service)
		}
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
		if isUnsupportedPath(err) {
			return nil, fmt.Errorf("the service dependencies endpoint is not exposed by this CubeAPM deployment")
		}
		return nil, fmt.Errorf("getting dependencies: %w", err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("dependencies error: %s", resp.Errors[0].Msg)
	}

	return resp.Data, nil
}
