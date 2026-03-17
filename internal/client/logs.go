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

	// Stats query returns newline-delimited JSON
	result := &types.StatsResult{}
	err = c.streamJSON(resp, func(raw json.RawMessage) error {
		var fields map[string]interface{}
		if err := json.Unmarshal(raw, &fields); err != nil {
			return fmt.Errorf("parsing stats row: %w", err)
		}
		row := types.StatsResultRow{
			Fields: make(map[string]string),
		}
		for k, v := range fields {
			row.Fields[k] = fmt.Sprintf("%v", v)
		}
		result.Rows = append(result.Rows, row)
		return nil
	})

	return result, err
}

// GetLogStreams returns log streams.
func (c *Client) GetLogStreams(query string, from, to time.Time) ([]types.StreamInfo, error) {
	params := url.Values{}
	if query != "" {
		params.Set("query", query)
	}
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

	var streams []types.StreamInfo
	err = c.streamJSON(resp, func(raw json.RawMessage) error {
		var s types.StreamInfo
		if err := json.Unmarshal(raw, &s); err != nil {
			return fmt.Errorf("parsing stream: %w", err)
		}
		streams = append(streams, s)
		return nil
	})

	return streams, err
}

// GetLogFieldNames returns available log field names.
func (c *Client) GetLogFieldNames(query string, from, to time.Time) ([]types.FieldInfo, error) {
	params := url.Values{}
	if query != "" {
		params.Set("query", query)
	}
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

	var fields []types.FieldInfo
	err = c.streamJSON(resp, func(raw json.RawMessage) error {
		var f types.FieldInfo
		if err := json.Unmarshal(raw, &f); err != nil {
			return fmt.Errorf("parsing field: %w", err)
		}
		fields = append(fields, f)
		return nil
	})

	return fields, err
}

// GetLogFieldValues returns values for a specific log field.
func (c *Client) GetLogFieldValues(field, query string, from, to time.Time, limit int) ([]types.FieldValueInfo, error) {
	params := url.Values{}
	params.Set("field", field)
	if query != "" {
		params.Set("query", query)
	}
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

	var values []types.FieldValueInfo
	err = c.streamJSON(resp, func(raw json.RawMessage) error {
		var v types.FieldValueInfo
		if err := json.Unmarshal(raw, &v); err != nil {
			return fmt.Errorf("parsing field value: %w", err)
		}
		values = append(values, v)
		return nil
	})

	return values, err
}
