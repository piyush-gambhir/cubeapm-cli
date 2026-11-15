package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code,omitempty"`
}

func WriteError(w io.Writer, format string, err error, statusCode int) {
	if format == "json" {
		resp := ErrorResponse{Error: err.Error(), StatusCode: statusCode}
		json.NewEncoder(w).Encode(resp)
	} else {
		fmt.Fprintf(w, "Error: %v\n", err)
	}
}
