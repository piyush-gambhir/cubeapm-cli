package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestWriteError_PlainText(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("something went wrong")

	WriteError(&buf, "table", err, 0)

	got := buf.String()
	if !strings.Contains(got, "Error: something went wrong") {
		t.Errorf("plain text error output = %q, want it to contain 'Error: something went wrong'", got)
	}
}

func TestWriteError_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("bad request")

	WriteError(&buf, "json", err, 400)

	var resp ErrorResponse
	if jsonErr := json.Unmarshal(buf.Bytes(), &resp); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", jsonErr, buf.String())
	}

	if resp.Error != "bad request" {
		t.Errorf("resp.Error = %q, want %q", resp.Error, "bad request")
	}
	if resp.StatusCode != 400 {
		t.Errorf("resp.StatusCode = %d, want %d", resp.StatusCode, 400)
	}
}

func TestWriteError_JSON_NoStatusCode(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("generic error")

	WriteError(&buf, "json", err, 0)

	var resp ErrorResponse
	if jsonErr := json.Unmarshal(buf.Bytes(), &resp); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v\nraw: %s", jsonErr, buf.String())
	}

	if resp.Error != "generic error" {
		t.Errorf("resp.Error = %q, want %q", resp.Error, "generic error")
	}

	// status_code should be omitted (zero value with omitempty)
	raw := buf.String()
	if strings.Contains(raw, "status_code") {
		t.Errorf("JSON output should omit status_code when 0, got: %s", raw)
	}
}

func TestWriteError_YAML_FallsBackToPlainText(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("yaml error")

	WriteError(&buf, "yaml", err, 0)

	got := buf.String()
	if !strings.Contains(got, "Error: yaml error") {
		t.Errorf("YAML format should fall back to plain text, got: %q", got)
	}
}
