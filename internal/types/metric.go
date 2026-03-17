package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// MetricsResult represents a Prometheus API response.
type MetricsResult struct {
	Status    string          `json:"status"`
	Data      MetricsData     `json:"data"`
	ErrorType string          `json:"errorType,omitempty"`
	Error     string          `json:"error,omitempty"`
	Warnings  []string        `json:"warnings,omitempty"`
}

// MetricsData contains the result type and actual data.
type MetricsData struct {
	ResultType string       `json:"resultType"` // vector, matrix, scalar, string
	Result     json.RawMessage `json:"result"`
}

// VectorResult is a list of instant-query samples.
type VectorResult []Sample

// MatrixResult is a list of range-query series.
type MatrixResult []Series

// Sample represents a single instant vector sample.
type Sample struct {
	Metric map[string]string `json:"metric"`
	Value  SamplePair        `json:"value"`
}

// Series represents a range vector result.
type Series struct {
	Metric map[string]string `json:"metric"`
	Values []SamplePair      `json:"values"`
}

// SamplePair is a [timestamp, value] pair from Prometheus.
type SamplePair [2]interface{}

// Timestamp returns the Unix timestamp from the sample pair.
func (s SamplePair) Timestamp() float64 {
	switch v := s[0].(type) {
	case float64:
		return v
	case json.Number:
		f, _ := v.Float64()
		return f
	default:
		return 0
	}
}

// Value returns the string value from the sample pair.
func (s SamplePair) Value() string {
	switch v := s[1].(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// PromLabelsResponse wraps /api/v1/labels response.
type PromLabelsResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

// PromLabelValuesResponse wraps /api/v1/label/{label}/values response.
type PromLabelValuesResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

// PromSeriesResponse wraps /api/v1/series response.
type PromSeriesResponse struct {
	Status string              `json:"status"`
	Data   []map[string]string `json:"data"`
}
