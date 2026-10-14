package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/cmdutil"
)

func TestCommandsResolve(t *testing.T) {
	// Invalid config and a release build catch accidental use of the root hooks.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("CUBEAPM_SERVER", "invalid.example")
	t.Setenv("CUBEAPM_READ_ONLY", "true")
	t.Setenv("CUBEAPM_NO_INPUT", "true")
	path := filepath.Join(dir, "cubeapm-cli", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("invalid: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	oldVersion, oldTerminal, oldFormat := Version, stderrIsTerminal, cmdutil.OutputFormat
	Version, stderrIsTerminal = "v1.0.0", func() bool { t.Fatal("resolve checked the terminal for updates"); return true }
	t.Cleanup(func() { Version, stderrIsTerminal, cmdutil.OutputFormat = oldVersion, oldTerminal, oldFormat })

	for _, tc := range []struct {
		name    string
		args    []string
		command string
		effect  string
		flags   map[string]string
		pos     []string
		wantErr bool
	}{
		{name: "read", args: []string{"traces", "services"}, command: "cubeapm traces services", effect: "read", flags: map[string]string{}, pos: []string{}},
		{name: "write with leading global flag", args: []string{"-o", "json", "ingest", "logs", "--file", "logs.jsonl"}, command: "cubeapm ingest logs", effect: "remote_write", flags: map[string]string{"output": "json", "file": "logs.jsonl"}, pos: []string{}},
		{name: "positional and separator", args: []string{"logs", "query", "--", "--literal"}, command: "cubeapm logs query", effect: "read", flags: map[string]string{}, pos: []string{"--literal"}},
		{name: "fresh flags", args: []string{"ingest", "logs"}, command: "cubeapm ingest logs", effect: "remote_write", flags: map[string]string{}, pos: []string{}},
		{name: "unknown", args: []string{"nonsense"}, wantErr: true},
		{name: "root", wantErr: true},
		{name: "group", args: []string{"traces"}, wantErr: true},
		{name: "invalid flag", args: []string{"traces", "services", "--nonsense"}, wantErr: true},
		{name: "invalid value", args: []string{"-o", "json", "ingest", "logs", "--query-port", "bad"}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			root := newRootCmd()
			root.SetOut(&out)
			root.SetErr(&bytes.Buffer{})
			root.SetArgs(append([]string{"-o", "yaml", "commands", "resolve", "--"}, tc.args...))
			err := root.Execute()
			if tc.wantErr {
				if err == nil || out.Len() != 0 {
					t.Fatalf("expected error with no stdout, got %v, %s", err, &out)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var got resolution
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatalf("expected JSON despite outer YAML flag: %v, %s", err, &out)
			}
			want := resolution{Command: tc.command, Effect: tc.effect, Flags: tc.flags, Args: tc.pos}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %+v, want %+v", got, want)
			}
		})
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "invalid: [" {
		t.Fatalf("resolve modified config: %s, %v", data, err)
	}
}

func TestCommandsResolveAnnotations(t *testing.T) {
	for _, tc := range []struct {
		name        string
		ann         map[string]string
		set         bool
		effect      string
		conditional bool
	}{
		{"remote", map[string]string{"mutates": "true", "writes-local": "true"}, false, "remote_write", false},
		{"remote conditional", map[string]string{"mutates": "conditional"}, false, "remote_write", true},
		{"remote flag absent", map[string]string{"mutates-when-flag": "write"}, false, "read", false},
		{"remote flag set false", map[string]string{"mutates-when-flag": "write"}, true, "remote_write", false},
		{"local", map[string]string{"writes-local": "true"}, false, "local_write", false},
		{"local conditional", map[string]string{"writes-local": "conditional"}, false, "local_write", true},
		{"self update", map[string]string{"self-update": "true"}, false, "local_write", false},
		{"local flag absent", map[string]string{"writes-local-when-flag": "write"}, false, "read", false},
		{"local flag set", map[string]string{"writes-local-when-flag": "write"}, true, "local_write", false},
		{"false annotations", map[string]string{"mutates": "false", "writes-local": "false"}, false, "read", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &cobra.Command{Use: "test", Annotations: tc.ann}
			target.Annotations["interactive"] = "true"
			target.Annotations["secret-output"] = "true"
			target.Flags().Bool("write", false, "")
			args := []string{}
			if tc.set {
				args = []string{"--write=false"}
			}
			if err := target.ParseFlags(args); err != nil {
				t.Fatal(err)
			}
			got := resolveCommand(target)
			if got.Effect != tc.effect || got.Conditional != tc.conditional || !got.Interactive || !got.Secret {
				t.Fatalf("unexpected resolution: %+v", got)
			}
		})
	}
}
