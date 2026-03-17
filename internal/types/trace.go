package types

// Trace represents a full trace in Jaeger format.
type Trace struct {
	TraceID   string             `json:"traceID"`
	Spans     []Span             `json:"spans"`
	Processes map[string]Process `json:"processes"`
	Warnings  []string           `json:"warnings,omitempty"`
}

// Span represents a single span in a trace.
type Span struct {
	TraceID       string     `json:"traceID"`
	SpanID        string     `json:"spanID"`
	OperationName string     `json:"operationName"`
	References    []SpanRef  `json:"references,omitempty"`
	StartTime     int64      `json:"startTime"` // microseconds since epoch
	Duration      int64      `json:"duration"`  // microseconds
	Tags          []KeyValue `json:"tags,omitempty"`
	Logs          []SpanLog  `json:"logs,omitempty"`
	ProcessID     string     `json:"processID"`
	Warnings      []string   `json:"warnings,omitempty"`
}

// SpanRef represents a reference to another span.
type SpanRef struct {
	RefType string `json:"refType"` // CHILD_OF or FOLLOWS_FROM
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

// KeyValue represents a key-value pair (tag or field).
type KeyValue struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// SpanLog represents a log entry within a span.
type SpanLog struct {
	Timestamp int64      `json:"timestamp"`
	Fields    []KeyValue `json:"fields"`
}

// Process represents the process/service associated with spans.
type Process struct {
	ServiceName string     `json:"serviceName"`
	Tags        []KeyValue `json:"tags,omitempty"`
}

// Operation represents a service operation.
type Operation struct {
	Name     string `json:"name"`
	SpanKind string `json:"spanKind,omitempty"`
}

// JaegerSearchResponse wraps the Jaeger search API response.
type JaegerSearchResponse struct {
	Data   []TraceSearchResult `json:"data"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors,omitempty"`
}

// JaegerTraceResponse wraps the Jaeger trace API response.
type JaegerTraceResponse struct {
	Data   []Trace `json:"data"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors,omitempty"`
}

// JaegerServicesResponse wraps the Jaeger services API response.
type JaegerServicesResponse struct {
	Data   []string `json:"data"`
	Total  int      `json:"total"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors,omitempty"`
}

// JaegerOperationsResponse wraps the Jaeger operations API response.
type JaegerOperationsResponse struct {
	Data   []Operation `json:"data"`
	Total  int         `json:"total"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors,omitempty"`
}

// JaegerDependenciesResponse wraps the Jaeger dependencies API response.
type JaegerDependenciesResponse struct {
	Data   []Dependency `json:"data"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors,omitempty"`
}
