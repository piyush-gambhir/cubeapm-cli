package types

// LogEntry represents a single log entry from VictoriaLogs.
type LogEntry struct {
	Message string            `json:"_msg"`
	Time    string            `json:"_time"`
	Stream  string            `json:"_stream"`
	Fields  map[string]string `json:"-"`
}

// StreamInfo represents a log stream.
type StreamInfo struct {
	Stream  string `json:"value"`
	Entries int64  `json:"hits"`
}

// FieldInfo represents a log field name and its hit count.
type FieldInfo struct {
	Name string `json:"value"`
	Hits int64  `json:"hits"`
}

// FieldValueInfo represents a log field value and its hit count.
type FieldValueInfo struct {
	Value string `json:"value"`
	Hits  int64  `json:"hits"`
}

// LogHit represents a single hit bucket in a log hits response.
type LogHit struct {
	Timestamps []string          `json:"timestamps,omitempty"`
	Values     []int64           `json:"values,omitempty"`
	Fields     map[string]string `json:"fields,omitempty"`
}

// LogHitsResult represents the /hits endpoint response.
type LogHitsResult struct {
	Hits []LogHit `json:"hits"`
}

// StatsResultRow represents a single row in stats query output.
type StatsResultRow struct {
	Fields map[string]string `json:"-"`
}

// StatsResult represents the /stats_query endpoint response.
type StatsResult struct {
	Rows []StatsResultRow `json:"rows,omitempty"`
}

// DeleteTask represents a log deletion task.
type DeleteTask struct {
	TaskID   string `json:"task_id"`
	Filter   string `json:"filter"`
	Status   string `json:"status"`
	Progress string `json:"progress"`
}

// DeleteRunResponse represents the response from starting a log delete.
type DeleteRunResponse struct {
	TaskID string `json:"task_id"`
}
