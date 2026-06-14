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

// JaegerSearchResponse wraps the Jaeger-format search API response.
// Classic Jaeger query servers return this shape; CubeAPM servers may
// return CubeSearchResultList instead. SearchTraces handles both.
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

// CubeSearchResultList is the native CubeAPM search response shape, a
// flat JSON array where each element wraps one trace with its "key span".
// The wire format differs from Jaeger in three ways: no outer data wrapper,
// snake_case span fields (trace_id/span_id/operation_name/start_time),
// and start_time as an RFC3339 string (not microseconds int64). See
// cubeToSearchResults for the conversion.
type CubeSearchResultList []CubeSearchResult

// CubeSearchResult wraps a single trace in the CubeAPM native format.
type CubeSearchResult struct {
	KeySpanID string    `json:"keySpanId"`
	Trace     CubeTrace `json:"trace"`
}

// CubeTrace is a trace as CubeAPM's query API returns it.
type CubeTrace struct {
	Spans     []CubeSpan         `json:"spans"`
	Processes map[string]Process `json:"processes,omitempty"`
}

// CubeSpan is a span with CubeAPM's snake_case field naming and an
// RFC3339-string start time. Tags are decoded generically because the
// wire format uses typed keys like v_str/v_int/v_bool rather than the
// generic {value} field Jaeger uses.
type CubeSpan struct {
	TraceID       string        `json:"trace_id"`
	SpanID        string        `json:"span_id"`
	OperationName string        `json:"operation_name"`
	References    []CubeSpanRef `json:"references,omitempty"`
	StartTime     string        `json:"start_time"`
	Duration      int64         `json:"duration"`
	Tags          []CubeSpanTag `json:"tags,omitempty"`
	ProcessID     string        `json:"process_id,omitempty"`
	// Process is the inline per-span process CubeAPM attaches in its native
	// responses (service_name + resource tags) when there is no top-level
	// process_map / process_id reference.
	Process *CubeProcess `json:"process,omitempty"`
}

// CubeProcess is CubeAPM's inline span process descriptor.
type CubeProcess struct {
	ServiceNameSnake string        `json:"service_name"`
	ServiceName      string        `json:"serviceName"`
	Tags             []CubeSpanTag `json:"tags,omitempty"`
}

// CubeSpanRef mirrors Jaeger's SpanRef but with snake_case on the wire.
type CubeSpanRef struct {
	RefType string `json:"ref_type"`
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
}

// CubeSpanTag is a single tag entry from CubeAPM's native format. Only
// one of the v_* fields is populated per entry.
type CubeSpanTag struct {
	Key    string  `json:"key"`
	VStr   string  `json:"v_str,omitempty"`
	VInt   int64   `json:"v_int,omitempty"`
	VBool  bool    `json:"v_bool,omitempty"`
	VFloat float64 `json:"v_float,omitempty"`
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
