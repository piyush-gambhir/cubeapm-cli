package logs

import "testing"

func TestPrependFilters(t *testing.T) {
	tests := []struct {
		name    string
		filters []string
		logsql  string
		want    string
	}{
		{
			name:    "no filters returns query unchanged",
			filters: nil,
			logsql:  "error OR timeout",
			want:    "error OR timeout",
		},
		{
			name:    "single filter parenthesizes user query",
			filters: []string{"service:api"},
			logsql:  "error",
			want:    "service:api AND (error)",
		},
		{
			name:    "OR query keeps its meaning under AND",
			filters: []string{"service:api"},
			logsql:  "error OR timeout",
			want:    "service:api AND (error OR timeout)",
		},
		{
			name:    "multiple filters joined with AND",
			filters: []string{"service:api", "level:error"},
			logsql:  "timeout",
			want:    "service:api AND level:error AND (timeout)",
		},
		{
			name:    "match-all query is not wrapped",
			filters: []string{"service:api"},
			logsql:  "*",
			want:    "service:api AND *",
		},
		{
			name:    "piped query is not wrapped",
			filters: []string{"service:api"},
			logsql:  "error | sort by (_time)",
			want:    "service:api AND error | sort by (_time)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prependFilters(tt.filters, tt.logsql)
			if got != tt.want {
				t.Errorf("prependFilters(%v, %q) = %q, want %q", tt.filters, tt.logsql, got, tt.want)
			}
		})
	}
}
