package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

const logsSelectBasePath = "/api/logs/select/logsql"

// QueryLogs queries logs using LogsQL syntax.
func (c *Client) QueryLogs(query string, from, to time.Time, limit int) ([]types.LogEntry, error) {
	var entries []types.LogEntry
	err := c.QueryLogsStream(query, from, to, limit, func(entry types.LogEntry) error {
		entries = append(entries, entry)
		return nil
	})
	return entries, err
}

// QueryLogsStream queries logs using LogsQL syntax with streaming output.
// It reads newline-delimited JSON and calls handler for each entry.
func (c *Client) QueryLogsStream(query string, from, to time.Time, limit int, handler func(types.LogEntry) error) error {
	params := url.Values{}
	params.Set("query", query)
	if !from.IsZero() {
		params.Set("start", from.Format(time.RFC3339Nano))
	}
	if !to.IsZero() {
		params.Set("end", to.Format(time.RFC3339Nano))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := c.post(c.queryBaseURL, logsSelectBasePath+"/query", params)
	if err != nil {
		return fmt.Errorf("querying logs: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return err
	}

	return c.streamJSON(resp, func(raw json.RawMessage) error {
		// Parse the raw JSON into a map first to capture all fields
		var fields map[string]interface{}
		if err := json.Unmarshal(raw, &fields); err != nil {
			return fmt.Errorf("parsing log entry: %w", err)
		}

		entry := types.LogEntry{
			Fields: make(map[string]string),
		}

		for k, v := range fields {
			strVal := fmt.Sprintf("%v", v)
			switch k {
			case "_msg":
				entry.Message = strVal
			case "_time":
				entry.Time = strVal
			case "_stream":
				entry.Stream = strVal
			default:
				entry.Fields[k] = strVal
			}
		}

		return handler(entry)
	})
}

// GetLogHits returns log volume/hits data.
func (c *Client) GetLogHits(query string, from, to time.Time, step string) (*types.LogHitsResult, error) {
	params := url.Values{}
	params.Set("query", query)
	if !from.IsZero() {
		params.Set("start", from.Format(time.RFC3339Nano))
	}
	if !to.IsZero() {
		params.Set("end", to.Format(time.RFC3339Nano))
	}
	if step != "" {
		params.Set("step", step)
	} else if !from.IsZero() && !to.IsZero() {
		// Without an explicit step the server collapses the whole range into a
		// single, mislabeled bucket. Derive ~60 buckets so the histogram and
		// per-bucket timestamps are meaningful by default.
		stepDur := to.Sub(from) / 60
		if stepDur < time.Second {
			stepDur = time.Second
		}
		params.Set("step", fmt.Sprintf("%ds", int(stepDur.Seconds())))
	}

	resp, err := c.get(c.queryBaseURL, logsSelectBasePath+"/hits", params)
	if err != nil {
		return nil, fmt.Errorf("getting log hits: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var result types.LogHitsResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing log hits: %w", err)
	}

	return &result, nil
}

// GetLogStats returns stats query results.
func (c *Client) GetLogStats(query string, from, to time.Time) (*types.StatsResult, error) {
	params := url.Values{}
	params.Set("query", query)
	if !from.IsZero() {
		params.Set("start", from.Format(time.RFC3339Nano))
	}
	if !to.IsZero() {
		params.Set("end", to.Format(time.RFC3339Nano))
	}

	resp, err := c.post(c.queryBaseURL, logsSelectBasePath+"/stats_query", params)
	if err != nil {
		return nil, fmt.Errorf("querying log stats: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	// CubeAPM returns Prometheus-style JSON for stats, not NDJSON:
	//   {"status":"success","data":{"resultType":"vector"|"matrix",
	//     "result":[{"metric":{labels},"value":[ts,"v"]      // vector
	//                              ,"values":[[ts,"v"],...]}]}} // matrix
	var promResp struct {
		Status string `json:"status"`
		Error  string `json:"error"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return nil, fmt.Errorf("parsing log stats: %w", err)
	}
	if promResp.Status != "" && promResp.Status != "success" {
		return nil, fmt.Errorf("stats query failed: %s", promResp.Error)
	}

	result := &types.StatsResult{}
	for _, r := range promResp.Data.Result {
		row := types.StatsResultRow{Fields: make(map[string]string)}
		for k, v := range r.Metric {
			row.Fields[k] = v
		}
		// The aggregated value is value[1] (vector) or the last of values[]
		// (matrix). Don't clobber a real label/alias literally named "value".
		if _, taken := row.Fields["value"]; !taken {
			if len(r.Value) == 2 {
				row.Fields["value"] = fmt.Sprintf("%v", r.Value[1])
			} else if n := len(r.Values); n > 0 && len(r.Values[n-1]) == 2 {
				row.Fields["value"] = fmt.Sprintf("%v", r.Values[n-1][1])
			}
		}
		result.Rows = append(result.Rows, row)
	}
	return result, nil
}

// GetLogStreams returns log streams.
//
// The server rejects empty queries with an opaque "missing query" error,
// so default to `*` (match-all) when the caller doesn't narrow the query.
// Callers that genuinely want no filter get what they'd expect anyway.
func (c *Client) GetLogStreams(query string, from, to time.Time) ([]types.StreamInfo, error) {
	params := url.Values{}
	if query == "" {
		query = "*"
	}
	params.Set("query", query)
	if !from.IsZero() {
		params.Set("start", from.Format(time.RFC3339Nano))
	}
	if !to.IsZero() {
		params.Set("end", to.Format(time.RFC3339Nano))
	}

	resp, err := c.get(c.queryBaseURL, logsSelectBasePath+"/streams", params)
	if err != nil {
		return nil, fmt.Errorf("getting log streams: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	// CubeAPM returns a single JSON object {"values":[{"value":..,"hits":..},...]},
	// not newline-delimited JSON.
	var wrapper struct {
		Values []types.StreamInfo `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("parsing log streams: %w", err)
	}
	return wrapper.Values, nil
}

// GetLogFieldNames returns available log field names.
//
// Default to `*` when the caller doesn't specify a query, the server
// rejects empty queries, and users generally want "all fields" when they
// don't narrow the search.
func (c *Client) GetLogFieldNames(query string, from, to time.Time) ([]types.FieldInfo, error) {
	params := url.Values{}
	if query == "" {
		query = "*"
	}
	params.Set("query", query)
	if !from.IsZero() {
		params.Set("start", from.Format(time.RFC3339Nano))
	}
	if !to.IsZero() {
		params.Set("end", to.Format(time.RFC3339Nano))
	}

	resp, err := c.get(c.queryBaseURL, logsSelectBasePath+"/field_names", params)
	if err != nil {
		return nil, fmt.Errorf("getting field names: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var wrapper struct {
		Values []types.FieldInfo `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("parsing field names: %w", err)
	}
	return wrapper.Values, nil
}

// GetLogFieldValues returns values for a specific log field.
func (c *Client) GetLogFieldValues(field, query string, from, to time.Time, limit int) ([]types.FieldValueInfo, error) {
	params := url.Values{}
	params.Set("field", field)
	if query == "" {
		query = "*"
	}
	params.Set("query", query)
	if !from.IsZero() {
		params.Set("start", from.Format(time.RFC3339Nano))
	}
	if !to.IsZero() {
		params.Set("end", to.Format(time.RFC3339Nano))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := c.get(c.queryBaseURL, logsSelectBasePath+"/field_values", params)
	if err != nil {
		return nil, fmt.Errorf("getting field values: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var wrapper struct {
		Values []types.FieldValueInfo `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("parsing field values: %w", err)
	}
	return wrapper.Values, nil
}
