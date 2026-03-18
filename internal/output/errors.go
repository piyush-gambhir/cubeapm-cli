package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// ErrorResponse represents a structured error for JSON output.
type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code,omitempty"`
}

// WriteError writes an error message to w. When format is "json", it outputs
// a structured JSON object. Otherwise it writes a plain-text error line.
func WriteError(w io.Writer, format string, err error, statusCode int) {
	if format == "json" {
		resp := ErrorResponse{Error: err.Error(), StatusCode: statusCode}
		json.NewEncoder(w).Encode(resp)
	} else {
		fmt.Fprintf(w, "Error: %v\n", err)
	}
}
