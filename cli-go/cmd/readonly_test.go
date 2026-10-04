package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

const readOnlyTestConfig = `current_profile: main
profiles:
  main:
    server: 127.0.0.1
    auth_method: none
    read_only: %READ_ONLY%
  other:
    server: 127.0.0.1
    auth_method: none
`

// runRoot executes the real root command against an isolated config file and
// returns the config file contents afterwards plus the command error.
func runRoot(t *testing.T, profileReadOnly bool, env map[string]string, args ...string) (string, error) {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	for _, name := range []string{"CUBEAPM_READ_ONLY", "CUBEAPM_SERVER", "CUBEAPM_NO_INPUT", "CUBEAPM_QUIET"} {
		t.Setenv(name, "")
	}
	for name, value := range env {
		t.Setenv(name, value)
	}

	path := filepath.Join(dir, "cubeapm-cli", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(testConfig(profileReadOnly)), 0o600); err != nil {
		t.Fatal(err)
	}

	// Cobra keeps flag values between Execute calls on the shared root.
	resetFlags := func() {
		for _, name := range []string{"read-only", "profile"} {
			f := rootCmd.PersistentFlags().Lookup(name)
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		}
		if update, _, err := rootCmd.Find([]string{"update"}); err == nil {
			check := update.Flags().Lookup("check")
			_ = check.Value.Set("false")
			check.Changed = false
		}
	}
	resetFlags()
	t.Cleanup(resetFlags)

	// Execute adds cobra's help and completion commands to the shared root;
	// remove them so the agent-safety manifest test sees the same graph.
	before := map[*cobra.Command]bool{}
	for _, c := range rootCmd.Commands() {
		before[c] = true
	}
	t.Cleanup(func() {
		for _, c := range rootCmd.Commands() {
			if !before[c] {
				rootCmd.RemoveCommand(c)
			}
		}
	})

	rootCmd.SetArgs(args)
	err := rootCmd.Execute()

	after, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(after), err
}

func testConfig(profileReadOnly bool) string {
	return strings.Replace(readOnlyTestConfig, "%READ_ONLY%", strconv.FormatBool(profileReadOnly), 1)
}

// configWriteCommands change the local config file.
var configWriteCommands = [][]string{
	{"config", "set", "server", "changed.example.com"},
	{"config", "profiles", "use", "other"},
	{"config", "profiles", "delete", "other"},
}

// writeCommands lists every command annotated mutates=true.
var writeCommands = append(append([][]string{}, configWriteCommands...), [][]string{
	{"update"},
	{"ingest", "metrics"},
	{"ingest", "logs"},
	{"logs", "delete", "run", "_time:<1h"},
	{"logs", "delete", "stop", "task-1"},
}...)

func TestReadOnlyBlocksEveryWriteCommand(t *testing.T) {
	sources := []struct {
		name    string
		profile bool
		env     map[string]string
		flags   []string
	}{
		{name: "flag", flags: []string{"--read-only"}},
		{name: "env", env: map[string]string{"CUBEAPM_READ_ONLY": "true"}},
		{name: "profile", profile: true},
	}

	for _, source := range sources {
		for _, command := range writeCommands {
			t.Run(source.name+"/"+strings.Join(command, " "), func(t *testing.T) {
				args := append(append([]string{}, source.flags...), command...)
				_, err := runRoot(t, source.profile, source.env, args...)
				if err == nil || !strings.Contains(err.Error(), "blocked in read-only mode") {
					t.Fatalf("expected read-only block, got %v", err)
				}
			})
		}
	}
}

func TestReadOnlyLeavesConfigUntouched(t *testing.T) {
	for _, command := range configWriteCommands {
		t.Run(strings.Join(command, " "), func(t *testing.T) {
			after, err := runRoot(t, false, nil, append([]string{"--read-only"}, command...)...)
			if err == nil {
				t.Fatal("expected read-only block")
			}
			if after != testConfig(false) {
				t.Fatalf("blocked command changed the config file:\n%s", after)
			}
		})
	}
}

// A profile's read_only: true cannot be lifted by the environment or the flag.
func TestReadOnlyProfileCannotBeDisabled(t *testing.T) {
	env := map[string]string{"CUBEAPM_READ_ONLY": "false"}
	_, err := runRoot(t, true, env, "--read-only=false", "config", "set", "server", "changed.example.com")
	if err == nil || !strings.Contains(err.Error(), "blocked in read-only mode") {
		t.Fatalf("expected profile read_only to stay on, got %v", err)
	}
}

func TestReadOnlyAllowsReads(t *testing.T) {
	if _, err := runRoot(t, true, nil, "config", "get", "server"); err != nil {
		t.Fatalf("config get should run in read-only mode: %v", err)
	}
	// update --check only reports; in tests it stops at the dev-build guard.
	_, err := runRoot(t, true, nil, "update", "--check")
	if err == nil || strings.Contains(err.Error(), "read-only") {
		t.Fatalf("update --check should pass the read-only check, got %v", err)
	}
}

func TestWriteCommandRunsWithoutReadOnly(t *testing.T) {
	after, err := runRoot(t, false, map[string]string{"CUBEAPM_READ_ONLY": "false"}, "config", "profiles", "use", "other")
	if err != nil {
		t.Fatalf("profiles use should run when read-only is off: %v", err)
	}
	if !strings.Contains(after, "current_profile: other") {
		t.Fatalf("profiles use did not switch the profile:\n%s", after)
	}
}

// Config writes and update cannot escape the active profile's read_only by
// selecting another (or a missing) profile with --profile.
func TestReadOnlyActiveProfileAppliesDespiteProfileFlag(t *testing.T) {
	for _, args := range [][]string{
		{"--profile", "other", "config", "set", "server", "changed.example.com"},
		{"--profile", "other", "config", "profiles", "delete", "main"},
		{"--profile", "missing", "update"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			after, err := runRoot(t, true, nil, args...)
			if err == nil || !strings.Contains(err.Error(), "blocked in read-only mode") {
				t.Fatalf("expected read-only block, got %v", err)
			}
			if after != testConfig(true) {
				t.Fatalf("blocked command changed the config file:\n%s", after)
			}
		})
	}
}
