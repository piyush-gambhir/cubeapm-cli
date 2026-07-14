package delete

import (
	"errors"
	"testing"
)

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"404 in message", errors.New("resource not found (HTTP 404): something"), true},
		{"not found in message", errors.New("task not found"), true},
		{"unrelated error", errors.New("connection refused"), false},
		{"403 error", errors.New("access denied (HTTP 403)"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNotFoundError(tt.err)
			if got != tt.want {
				t.Errorf("isNotFoundError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
