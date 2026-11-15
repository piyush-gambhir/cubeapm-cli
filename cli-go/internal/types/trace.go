package types

type Trace struct {
	TraceID   string             `json:"traceID"`
	Spans     []Span             `json:"spans"`
	Processes map[string]Process `json:"processes"`
	Warnings  []string           `json:"warnings,omitempty"`
}

type Span struct {
	TraceID       string     `json:"traceID"`
	SpanID        string     `json:"spanID"`
	OperationName string     `json:"operationName"`
	References    []SpanRef  `json:"references,omitempty"`
	StartTime     int64      `json:"startTime"`
	Duration      int64      `json:"duration"`
	Tags          []KeyValue `json:"tags,omitempty"`
	Logs          []SpanLog  `json:"logs,omitempty"`
	ProcessID     string     `json:"processID"`
	Warnings      []string   `json:"warnings,omitempty"`
}

type SpanRef struct {
	RefType string `json:"refType"`
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

type KeyValue struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type SpanLog struct {
	Timestamp int64      `json:"timestamp"`
	Fields    []KeyValue `json:"fields"`
}

type Process struct {
	ServiceName string     `json:"serviceName"`
	Tags        []KeyValue `json:"tags,omitempty"`
}

type Operation struct {
	Name     string `json:"name"`
	SpanKind string `json:"spanKind,omitempty"`
}

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

type CubeSearchResultList []CubeSearchResult

type CubeSearchResult struct {
	KeySpanID string    `json:"keySpanId"`
	Trace     CubeTrace `json:"trace"`
}

type CubeTrace struct {
	Spans     []CubeSpan         `json:"spans"`
	Processes map[string]Process `json:"processes,omitempty"`
}

type CubeSpan struct {
	TraceID       string        `json:"trace_id"`
	SpanID        string        `json:"span_id"`
	OperationName string        `json:"operation_name"`
	References    []CubeSpanRef `json:"references,omitempty"`
	StartTime     string        `json:"start_time"`
	Duration      int64         `json:"duration"`
	Tags          []CubeSpanTag `json:"tags,omitempty"`
	ProcessID     string        `json:"process_id,omitempty"`
	Process       *CubeProcess  `json:"process,omitempty"`
}

type CubeProcess struct {
	ServiceNameSnake string        `json:"service_name"`
	ServiceName      string        `json:"serviceName"`
	Tags             []CubeSpanTag `json:"tags,omitempty"`
}

type CubeSpanRef struct {
	RefType string `json:"ref_type"`
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
}

type CubeSpanTag struct {
	Key    string   `json:"key"`
	VStr   *string  `json:"v_str,omitempty"`
	VInt   *int64   `json:"v_int,omitempty"`
	VBool  *bool    `json:"v_bool,omitempty"`
	VFloat *float64 `json:"v_float,omitempty"`
}

type JaegerTraceResponse struct {
	Data   []Trace `json:"data"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors,omitempty"`
}

type JaegerServicesResponse struct {
	Data   []string `json:"data"`
	Total  int      `json:"total"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors,omitempty"`
}

type JaegerOperationsResponse struct {
	Data   []Operation `json:"data"`
	Total  int         `json:"total"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors,omitempty"`
}

type JaegerDependenciesResponse struct {
	Data   []Dependency `json:"data"`
	Errors []struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"errors,omitempty"`
}
