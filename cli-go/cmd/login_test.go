package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/config"
)

func TestLoginKeepsReadOnlyAndOutput(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"success","data":[]}`)
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := filepath.Join(dir, "cubeapm-cli", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	existing := "current_profile: prod\nprofiles:\n  prod:\n    server: old.example.com\n    output: json\n    read_only: true\n"
	if err := os.WriteFile(path, []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}

	stdin, err := os.CreateTemp(dir, "stdin")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(stdin, "prod\n%s\n2\n%s\n\n\n", ts.URL, u.Port())
	if _, err := stdin.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = stdin
	cmdutil.Quiet = true
	t.Cleanup(func() {
		os.Stdin = oldStdin
		cmdutil.Quiet = false
		cmdutil.AppConfig = nil
	})

	cmd := newLoginCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("login: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	p := cfg.Profiles["prod"]
	if p.Server != ts.URL {
		t.Errorf("server = %q, want %q", p.Server, ts.URL)
	}
	if !p.ReadOnly {
		t.Error("login removed read_only from the profile")
	}
	if p.Output != "json" {
		t.Errorf("output = %q, want json", p.Output)
	}
}
