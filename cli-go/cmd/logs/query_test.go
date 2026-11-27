package logs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/output"
)

func TestQueryYAMLStreamIsValid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"_msg":"first","_time":"2024-01-15T10:00:00Z","_stream":"{}","service":"api"}`)
		fmt.Fprintln(w, `{"_msg":"second","_time":"2024-01-15T10:01:00Z","_stream":"{}","service":"api"}`)
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	cmdutil.APIClient, err = client.NewClient(config.ResolvedConfig{Server: ts.URL, QueryPort: port, AuthMethod: "none"})
	if err != nil {
		t.Fatal(err)
	}
	cmdutil.OutputFormat = output.FormatYAML
	defer func() {
		cmdutil.APIClient = nil
		cmdutil.OutputFormat = ""
	}()

	var out bytes.Buffer
	cmd := newQueryCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"*"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("logs query: %v", err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(out.Bytes()))
	var msgs []string
	for {
		var doc map[string]string
		err := dec.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("invalid YAML stream: %v\n%s", err, out.String())
		}
		msgs = append(msgs, doc["_msg"])
	}
	if len(msgs) != 2 || msgs[0] != "first" || msgs[1] != "second" {
		t.Fatalf("got documents %q, want [first second]\n%s", msgs, out.String())
	}
}

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
			name:    "pipe inside a subquery is not a split point",
			filters: []string{"service:api"},
			logsql:  "user_id:in(_time:1h error | fields user_id)",
			want:    "service:api AND (user_id:in(_time:1h error | fields user_id))",
		},
		{
			name:    "half-open range keeps the split at depth zero",
			filters: []string{"service:api"},
			logsql:  "status:range[400, 500) | stats count()",
			want:    "service:api AND (status:range[400, 500)) | stats count()",
		},
		{
			name:    "comment line is skipped and the wrapper closes on a new line",
			filters: []string{"service:api"},
			logsql:  "_time:5m\n# per host | not a pipe\n| stats by (host) count()",
			want:    "service:api AND (_time:5m\n# per host | not a pipe\n) | stats by (host) count()",
		},
		{
			name:    "hash and pipe inside quotes are literal",
			filters: []string{"service:api"},
			logsql:  "\"a#b|c\" OR 'it\\'s|x' OR `raw|y` | limit 5",
			want:    "service:api AND (\"a#b|c\" OR 'it\\'s|x' OR `raw|y`) | limit 5",
		},
		{
			name:    "piped query wraps only the filter part",
			filters: []string{"service:api"},
			logsql:  "error OR timeout | sort by (_time)",
			want:    "service:api AND (error OR timeout) | sort by (_time)",
		},
		{
			name:    "pipe inside a quoted phrase is not a pipe",
			filters: []string{"service:api"},
			logsql:  `"a|b" OR timeout`,
			want:    `service:api AND ("a|b" OR timeout)`,
		},
		{
			name:    "match-all with pipes is not wrapped",
			filters: []string{"service:api"},
			logsql:  "* | stats count() as c",
			want:    "service:api AND * | stats count() as c",
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
