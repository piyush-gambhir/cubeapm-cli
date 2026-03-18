package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONFormatter(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	data := map[string]string{"name": "test", "value": "123"}
	err := f.Format(&buf, data)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"name"`) {
		t.Errorf("output missing 'name' key: %s", output)
	}
	if !strings.Contains(output, `"test"`) {
		t.Errorf("output missing 'test' value: %s", output)
	}

	// Verify it's valid JSON
	var parsed map[string]string
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}

	// Test with TableDef input -- should convert to list of maps
	var buf2 bytes.Buffer
	td := TableDef{
		Headers: []string{"Name", "Value"},
		Rows: [][]string{
			{"alpha", "1"},
			{"beta", "2"},
		},
	}
	err = f.Format(&buf2, td)
	if err != nil {
		t.Fatalf("Format(TableDef) error = %v", err)
	}

	var parsedRows []map[string]string
	if err := json.Unmarshal(buf2.Bytes(), &parsedRows); err != nil {
		t.Fatalf("TableDef JSON output is not valid: %v", err)
	}
	if len(parsedRows) != 2 {
		t.Errorf("got %d rows, want 2", len(parsedRows))
	}
	if parsedRows[0]["Name"] != "alpha" {
		t.Errorf("parsedRows[0][Name] = %q, want %q", parsedRows[0]["Name"], "alpha")
	}
}

func TestYAMLFormatter(t *testing.T) {
	f := &YAMLFormatter{}
	var buf bytes.Buffer

	data := map[string]string{"name": "test", "value": "123"}
	err := f.Format(&buf, data)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name:") {
		t.Errorf("output missing 'name' key: %s", output)
	}
	if !strings.Contains(output, "test") {
		t.Errorf("output missing 'test' value: %s", output)
	}
	if !strings.Contains(output, "value:") {
		t.Errorf("output missing 'value' key: %s", output)
	}
}

func TestTableFormatter(t *testing.T) {
	f := &TableFormatter{NoColor: true}
	var buf bytes.Buffer

	td := TableDef{
		Headers: []string{"SERVICE", "STATUS"},
		Rows: [][]string{
			{"api-gateway", "running"},
			{"payments", "stopped"},
		},
	}

	err := f.Format(&buf, td)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "SERVICE") {
		t.Errorf("output missing header 'SERVICE': %s", output)
	}
	if !strings.Contains(output, "STATUS") {
		t.Errorf("output missing header 'STATUS': %s", output)
	}
	if !strings.Contains(output, "api-gateway") {
		t.Errorf("output missing 'api-gateway': %s", output)
	}
	if !strings.Contains(output, "payments") {
		t.Errorf("output missing 'payments': %s", output)
	}

	// Verify no ANSI escape codes when NoColor=true
	if strings.Contains(output, "\033[") {
		t.Errorf("output contains ANSI escape codes when NoColor=true: %s", output)
	}

	// Test with color enabled
	var buf2 bytes.Buffer
	fColor := &TableFormatter{NoColor: false}
	err = fColor.Format(&buf2, td)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	colorOutput := buf2.String()
	if !strings.Contains(colorOutput, "\033[") {
		t.Errorf("output missing ANSI escape codes when NoColor=false: %s", colorOutput)
	}

	// Test with non-TableDef falls back to JSON
	var buf3 bytes.Buffer
	err = f.Format(&buf3, map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if !strings.Contains(buf3.String(), `"key"`) {
		t.Errorf("non-TableDef data did not fall back to JSON: %s", buf3.String())
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"json", FormatJSON, false},
		{"yaml", FormatYAML, false},
		{"table", FormatTable, false},
		{"", FormatTable, false}, // empty defaults to table
		{"xml", "", true},
		{"CSV", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewFormatter_Default(t *testing.T) {
	// Default (table) format
	f := NewFormatter(FormatTable, false)
	if _, ok := f.(*TableFormatter); !ok {
		t.Errorf("NewFormatter(FormatTable) returned %T, want *TableFormatter", f)
	}

	// JSON format
	f = NewFormatter(FormatJSON, false)
	if _, ok := f.(*JSONFormatter); !ok {
		t.Errorf("NewFormatter(FormatJSON) returned %T, want *JSONFormatter", f)
	}

	// YAML format
	f = NewFormatter(FormatYAML, false)
	if _, ok := f.(*YAMLFormatter); !ok {
		t.Errorf("NewFormatter(FormatYAML) returned %T, want *YAMLFormatter", f)
	}

	// Unknown format defaults to table
	f = NewFormatter("unknown", false)
	if _, ok := f.(*TableFormatter); !ok {
		t.Errorf("NewFormatter(unknown) returned %T, want *TableFormatter", f)
	}
}
