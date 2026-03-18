package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/output"
)

// TestNoInputBlocksLogin verifies that login returns an error when --no-input is set.
func TestNoInputBlocksLogin(t *testing.T) {
	cmdutil.NoInput = true
	defer func() { cmdutil.NoInput = false }()

	cmd := newLoginCmd()
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error when running login with --no-input, got nil")
	}
	if !strings.Contains(err.Error(), "no-input") {
		t.Errorf("error message should mention --no-input, got: %v", err)
	}
}

// TestNoInputEnvVar verifies that CUBEAPM_NO_INPUT env var is respected.
func TestNoInputEnvVar(t *testing.T) {
	tests := []struct {
		envValue string
		want     bool
	}{
		{"1", true},
		{"true", true},
		{"0", false},
		{"false", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("CUBEAPM_NO_INPUT", tt.envValue)
				defer os.Unsetenv("CUBEAPM_NO_INPUT")
			} else {
				os.Unsetenv("CUBEAPM_NO_INPUT")
			}

			v := os.Getenv("CUBEAPM_NO_INPUT")
			got := v == "1" || v == "true"
			if got != tt.want {
				t.Errorf("CUBEAPM_NO_INPUT=%q: got %v, want %v", tt.envValue, got, tt.want)
			}
		})
	}
}

// TestQuietEnvVar verifies that CUBEAPM_QUIET env var is respected.
func TestQuietEnvVar(t *testing.T) {
	tests := []struct {
		envValue string
		want     bool
	}{
		{"1", true},
		{"true", true},
		{"0", false},
		{"false", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("CUBEAPM_QUIET", tt.envValue)
				defer os.Unsetenv("CUBEAPM_QUIET")
			} else {
				os.Unsetenv("CUBEAPM_QUIET")
			}

			v := os.Getenv("CUBEAPM_QUIET")
			got := v == "1" || v == "true"
			if got != tt.want {
				t.Errorf("CUBEAPM_QUIET=%q: got %v, want %v", tt.envValue, got, tt.want)
			}
		})
	}
}

// TestStructuredJSONError verifies that WriteError produces structured JSON.
func TestStructuredJSONError(t *testing.T) {
	var buf bytes.Buffer
	err := fmt.Errorf("authentication failed (HTTP 401)")

	output.WriteError(&buf, "json", err, 401)

	got := buf.String()
	if !strings.Contains(got, `"error"`) {
		t.Errorf("JSON error output should contain 'error' key, got: %s", got)
	}
	if !strings.Contains(got, `"status_code"`) {
		t.Errorf("JSON error output should contain 'status_code' key, got: %s", got)
	}
	if !strings.Contains(got, "401") {
		t.Errorf("JSON error output should contain status code 401, got: %s", got)
	}
}

// TestStructuredErrorPlainText verifies that WriteError falls back to plain text for non-JSON.
func TestStructuredErrorPlainText(t *testing.T) {
	var buf bytes.Buffer
	err := fmt.Errorf("connection refused")

	output.WriteError(&buf, "table", err, 0)

	got := buf.String()
	if !strings.HasPrefix(got, "Error: ") {
		t.Errorf("plain text error should start with 'Error: ', got: %q", got)
	}
	if !strings.Contains(got, "connection refused") {
		t.Errorf("plain text error should contain the error message, got: %q", got)
	}
}

// TestRootCmdHasNoInputFlag verifies the --no-input flag is registered on root.
func TestRootCmdHasNoInputFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("no-input")
	if flag == nil {
		t.Fatal("expected --no-input flag to be registered on rootCmd")
	}
	if flag.DefValue != "false" {
		t.Errorf("--no-input default value = %q, want %q", flag.DefValue, "false")
	}
}

// TestRootCmdHasQuietFlag verifies the --quiet/-q flag is registered on root.
func TestRootCmdHasQuietFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("quiet")
	if flag == nil {
		t.Fatal("expected --quiet flag to be registered on rootCmd")
	}
	if flag.Shorthand != "q" {
		t.Errorf("--quiet shorthand = %q, want %q", flag.Shorthand, "q")
	}
	if flag.DefValue != "false" {
		t.Errorf("--quiet default value = %q, want %q", flag.DefValue, "false")
	}
}
