package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type MetricsResult struct {
	Status    string      `json:"status"`
	Data      MetricsData `json:"data"`
	ErrorType string      `json:"errorType,omitempty"`
	Error     string      `json:"error,omitempty"`
	Warnings  []string    `json:"warnings,omitempty"`
}

type MetricsData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

type VectorResult []Sample

type MatrixResult []Series

type Sample struct {
	Metric map[string]string `json:"metric"`
	Value  SamplePair        `json:"value"`
}

type Series struct {
	Metric map[string]string `json:"metric"`
	Values []SamplePair      `json:"values"`
}

type SamplePair [2]interface{}

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

type PromLabelsResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

type PromLabelValuesResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

type PromSeriesResponse struct {
	Status string              `json:"status"`
	Data   []map[string]string `json:"data"`
}
