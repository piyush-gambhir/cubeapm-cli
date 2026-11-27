package types

type LogEntry struct {
	Message string            `json:"_msg"`
	Time    string            `json:"_time"`
	Stream  string            `json:"_stream"`
	Fields  map[string]string `json:"-"`
}

type StreamInfo struct {
	Stream  string `json:"value"`
	Entries int64  `json:"hits"`
}

type FieldInfo struct {
	Name string `json:"value"`
	Hits int64  `json:"hits"`
}

type FieldValueInfo struct {
	Value string `json:"value"`
	Hits  int64  `json:"hits"`
}

type LogHit struct {
	Timestamps []string          `json:"timestamps,omitempty"`
	Values     []int64           `json:"values,omitempty"`
	Fields     map[string]string `json:"fields,omitempty"`
}

type LogHitsResult struct {
	Hits []LogHit `json:"hits"`
}

type StatsResultRow struct {
	Fields map[string]string `json:"-"`
}

type StatsResult struct {
	Rows []StatsResultRow `json:"rows,omitempty"`
}

type DeleteTask struct {
	TaskID    string `json:"task_id"`
	Filter    string `json:"filter"`
	StartTime string `json:"start_time"`
}

type DeleteRunResponse struct {
	TaskID string `json:"task_id"`
}
