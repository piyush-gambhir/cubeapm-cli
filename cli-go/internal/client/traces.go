package client

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/types"
)

func isUnsupportedPath(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unsupported path")
}

const tracesBasePath = "/api/traces/api/v1"

func (c *Client) SearchTraces(service string, tags map[string]string, query string, from, to time.Time, limit int, spanKind, minDuration, maxDuration, index string) ([]types.TraceSearchResult, error) {
	minDur, err := parseDurationFilter("minimum", minDuration)
	if err != nil {
		return nil, err
	}
	maxDur, err := parseDurationFilter("maximum", maxDuration)
	if err != nil {
		return nil, err
	}
	if minDur > 0 && maxDur > 0 && minDur > maxDur {
		return nil, fmt.Errorf("minimum duration %s is greater than maximum duration %s", minDur, maxDur)
	}
	tags = maps.Clone(tags)
	params := url.Values{}
	if index != "" {
		params.Set("index", index)
	}
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
		params.Set("query", query)
	}
	if !from.IsZero() {
		params.Set("start", strconv.FormatInt(from.Unix(), 10))
	}
	if !to.IsZero() {
		params.Set("end", strconv.FormatInt(to.Unix(), 10))
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
	if minDur > 0 && maxDur == 0 && strings.EqualFold(spanKind, "server") {
		params.Set("sortBy", "duration")
	}

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
	return filterByRootDuration(cubeToSearchResults(native), minDur, maxDur), nil
}

func parseDurationFilter(bound, value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil || d < 0 {
		return 0, fmt.Errorf("invalid %s duration %q: use a Go duration such as 500ms, 2s or 1m", bound, value)
	}
	return d, nil
}

func filterByRootDuration(results []types.TraceSearchResult, minDur, maxDur time.Duration) []types.TraceSearchResult {
	if minDur == 0 && maxDur == 0 {
		return results
	}
	out := make([]types.TraceSearchResult, 0, len(results))
	for _, r := range results {
		d := time.Duration(types.RootSpan(r.Spans).Duration) * time.Microsecond
		if minDur > 0 && d < minDur {
			continue
		}
		if maxDur > 0 && d > maxDur {
			continue
		}
		out = append(out, r)
	}
	return out
}

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
			Duration:      time.Duration(cs.Duration).Microseconds(),
			Tags:          convertCubeTags(cs.Tags),
			ProcessID:     pid,
		})
	}
	return spans, procs, traceID
}

func decodeCubeID(s string) string {
	if s == "" {
		return ""
	}
	if len(s)%2 == 0 {
		if b, err := hex.DecodeString(s); err == nil && len(b) > 0 {
			return hex.EncodeToString(b)
		}
	}
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.RawURLEncoding} {
		if b, err := enc.DecodeString(s); err == nil && len(b) > 0 {
			return hex.EncodeToString(b)
		}
	}
	return s
}

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
		refType := r.RefType
		if refType == "" {
			refType = "CHILD_OF"
		}
		out = append(out, types.SpanRef{
			RefType: refType,
			TraceID: decodeCubeID(r.TraceID),
			SpanID:  decodeCubeID(r.SpanID),
		})
	}
	return out
}

func convertCubeTags(in []types.CubeSpanTag) []types.KeyValue {
	if len(in) == 0 {
		return nil
	}
	out := make([]types.KeyValue, 0, len(in))
	for _, t := range in {
		kv := types.KeyValue{Key: t.Key}
		switch {
		case t.VStr != nil:
			kv.Type = "string"
			kv.Value = *t.VStr
		case t.VInt != nil:
			kv.Type = "int64"
			kv.Value = *t.VInt
		case t.VBool != nil:
			kv.Type = "bool"
			kv.Value = *t.VBool
		case t.VFloat != nil:
			kv.Type = "float64"
			kv.Value = *t.VFloat
		default:
			kv.Type = "string"
			kv.Value = ""
		}
		out = append(out, kv)
	}
	return out
}

func (c *Client) getRaw(baseURL, path string, params url.Values) ([]byte, error) {
	resp, err := c.get(baseURL, path, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}
	return readAllLimited(resp.Body)
}

func (c *Client) GetTrace(traceID string, from, to time.Time) (*types.Trace, error) {
	params := url.Values{}
	if !from.IsZero() {
		params.Set("start", strconv.FormatInt(from.Unix(), 10))
	}
	if !to.IsZero() {
		params.Set("end", strconv.FormatInt(to.Unix(), 10))
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
	if len(dual.Data) > 0 {
		return &dual.Data[0], nil
	}
	if len(dual.Spans) > 0 {
		return cubeSpansToTrace(traceID, dual.Spans, dual.ProcessMap, dual.Processes), nil
	}
	return nil, fmt.Errorf("trace %s not found in the searched time range (default: the last hour); widen it with --last or --from/--to", traceID)
}

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
	var obj map[string]proc
	if err := json.Unmarshal(raw, &obj); err == nil {
		for id, p := range obj {
			out[id] = pick(p)
		}
	}
	return out
}

func (c *Client) GetServices() ([]string, error) {
	var resp types.JaegerServicesResponse
	jaegerErr := c.getJSON(c.queryBaseURL, tracesBasePath+"/services", nil, &resp)
	if jaegerErr == nil {
		if len(resp.Errors) > 0 {
		} else {
			return resp.Data, nil
		}
	}

	services, fbErr := c.GetLabelValues("service", nil, time.Time{}, time.Time{})
	if fbErr == nil && len(services) > 0 {
		return services, nil
	}

	if jaegerErr != nil {
		return nil, fmt.Errorf("getting services: %w", jaegerErr)
	}
	return nil, fmt.Errorf("services error: %s", resp.Errors[0].Msg)
}

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
