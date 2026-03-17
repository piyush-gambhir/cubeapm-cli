package types

import (
	"fmt"
	"time"
)

// TimeRange represents a time range for queries.
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// APIError represents an error returned by the CubeAPM API.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	ErrorType  string `json:"errorType,omitempty"`
	Error      string `json:"error,omitempty"`
}

func (e *APIError) String() string {
	if e.Error != "" {
		return fmt.Sprintf("API error (HTTP %d): %s: %s", e.StatusCode, e.ErrorType, e.Error)
	}
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// Dependency represents a service dependency link.
type Dependency struct {
	Parent    string `json:"parent"`
	Child     string `json:"child"`
	CallCount int64  `json:"callCount"`
}

// TraceSearchResult represents a trace returned from a search query.
type TraceSearchResult struct {
	TraceID   string  `json:"traceID"`
	Spans     []Span  `json:"spans"`
	Processes map[string]Process `json:"processes"`
}
