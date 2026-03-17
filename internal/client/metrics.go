package client

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

const metricsBasePath = "/api/metrics/api/v1"

// QueryInstant performs an instant PromQL query.
func (c *Client) QueryInstant(query string, t time.Time) (*types.MetricsResult, error) {
	params := url.Values{}
	params.Set("query", query)
	if !t.IsZero() {
		params.Set("time", strconv.FormatFloat(float64(t.Unix())+float64(t.Nanosecond())/1e9, 'f', 3, 64))
	}

	var resp types.MetricsResult
	if err := c.postJSON(c.queryBaseURL, metricsBasePath+"/query", params, &resp); err != nil {
		return nil, fmt.Errorf("instant query: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("query error (%s): %s", resp.ErrorType, resp.Error)
	}

	return &resp, nil
}

// QueryRange performs a range PromQL query.
func (c *Client) QueryRange(query string, from, to time.Time, step string) (*types.MetricsResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatFloat(float64(from.Unix())+float64(from.Nanosecond())/1e9, 'f', 3, 64))
	params.Set("end", strconv.FormatFloat(float64(to.Unix())+float64(to.Nanosecond())/1e9, 'f', 3, 64))
	if step != "" {
		params.Set("step", step)
	} else {
		// Auto-calculate step: aim for ~250 data points
		duration := to.Sub(from)
		stepDur := duration / 250
		if stepDur < time.Second {
			stepDur = time.Second
		}
		params.Set("step", fmt.Sprintf("%ds", int(stepDur.Seconds())))
	}

	var resp types.MetricsResult
	if err := c.postJSON(c.queryBaseURL, metricsBasePath+"/query_range", params, &resp); err != nil {
		return nil, fmt.Errorf("range query: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("query error (%s): %s", resp.ErrorType, resp.Error)
	}

	return &resp, nil
}

// GetLabels returns all label names.
func (c *Client) GetLabels(from, to time.Time) ([]string, error) {
	params := url.Values{}
	if !from.IsZero() {
		params.Set("start", strconv.FormatFloat(float64(from.Unix()), 'f', 0, 64))
	}
	if !to.IsZero() {
		params.Set("end", strconv.FormatFloat(float64(to.Unix()), 'f', 0, 64))
	}

	var resp types.PromLabelsResponse
	if err := c.getJSON(c.queryBaseURL, metricsBasePath+"/labels", params, &resp); err != nil {
		return nil, fmt.Errorf("getting labels: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("labels query failed")
	}

	return resp.Data, nil
}

// GetLabelValues returns values for a specific label.
func (c *Client) GetLabelValues(label string, from, to time.Time) ([]string, error) {
	params := url.Values{}
	if !from.IsZero() {
		params.Set("start", strconv.FormatFloat(float64(from.Unix()), 'f', 0, 64))
	}
	if !to.IsZero() {
		params.Set("end", strconv.FormatFloat(float64(to.Unix()), 'f', 0, 64))
	}

	var resp types.PromLabelValuesResponse
	path := fmt.Sprintf("%s/label/%s/values", metricsBasePath, url.PathEscape(label))
	if err := c.getJSON(c.queryBaseURL, path, params, &resp); err != nil {
		return nil, fmt.Errorf("getting label values for %s: %w", label, err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("label values query failed")
	}

	return resp.Data, nil
}

// GetSeries returns matching time series.
func (c *Client) GetSeries(match []string, from, to time.Time, limit int) ([]map[string]string, error) {
	params := url.Values{}
	for _, m := range match {
		params.Add("match[]", m)
	}
	if !from.IsZero() {
		params.Set("start", strconv.FormatFloat(float64(from.Unix()), 'f', 0, 64))
	}
	if !to.IsZero() {
		params.Set("end", strconv.FormatFloat(float64(to.Unix()), 'f', 0, 64))
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	var resp types.PromSeriesResponse
	if err := c.getJSON(c.queryBaseURL, metricsBasePath+"/series", params, &resp); err != nil {
		return nil, fmt.Errorf("getting series: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("series query failed")
	}

	return resp.Data, nil
}

// FormatMetricLabels formats a metric label set as a string like {label="value",...}.
func FormatMetricLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}

	parts := make([]string, 0, len(labels))
	// Put __name__ first if present
	if name, ok := labels["__name__"]; ok {
		parts = append(parts, name)
	}

	labelParts := make([]string, 0, len(labels))
	for k, v := range labels {
		if k == "__name__" {
			continue
		}
		labelParts = append(labelParts, fmt.Sprintf(`%s="%s"`, k, v))
	}

	if len(labelParts) > 0 {
		parts = append(parts, "{"+strings.Join(labelParts, ", ")+"}")
	}

	return strings.Join(parts, "")
}
