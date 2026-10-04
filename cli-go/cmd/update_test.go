package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/internal/update"
)

// updateEnv runs the real root command as release 0.2.9 against a fake
// GitHub, with stderr and stdin treated as terminals and every opt-out unset.
type updateEnv struct {
	t            *testing.T
	configDir    string
	lookups      atomic.Int32
	downloadHits atomic.Int32
	notice       bytes.Buffer
	out          bytes.Buffer
}

func newUpdateEnv(t *testing.T, latestTag string, block <-chan struct{}) *updateEnv {
	t.Helper()
	env := &updateEnv{t: t}

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+updateRepo+"/releases/latest" {
			env.downloadHits.Add(1)
			http.NotFound(w, r)
			return
		}
		env.lookups.Add(1)
		if block != nil {
			<-block
		}
		w.Header().Set("Location", srv.URL+"/"+updateRepo+"/releases/tag/"+latestTag)
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	origBase := update.GitHubBaseURL
	origVersion, origStderr, origStdin, origNotice := Version, stderrIsTerminal, stdinIsTerminal, noticeOutput
	update.GitHubBaseURL = srv.URL
	Version = "0.2.9"
	stderrIsTerminal = func() bool { return true }
	stdinIsTerminal = func() bool { return true }
	noticeOutput = &env.notice
	pendingUpdateCheck = nil
	rootCmd.SetOut(&env.out)
	t.Cleanup(func() {
		update.GitHubBaseURL = origBase
		Version, stderrIsTerminal, stdinIsTerminal, noticeOutput = origVersion, origStderr, origStdin, origNotice
		pendingUpdateCheck = nil
		rootCmd.SetOut(nil)
	})

	for _, name := range []string{"CI", "CUBEAPM_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER", "CUBEAPM_QUIET",
		"CUBEAPM_NO_INPUT", "CUBEAPM_READ_ONLY", "CUBEAPM_SERVER", "GOBIN"} {
		t.Setenv(name, "")
	}
	// Keep the test binary out of any Go bin directory unless a test opts in.
	t.Setenv("GOPATH", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	env.configDir = filepath.Join(xdg, "cubeapm-cli")
	if err := os.MkdirAll(env.configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(env.configDir, "config.yaml"), []byte(testConfig(false)), 0o600); err != nil {
		t.Fatal(err)
	}
	return env
}

func (e *updateEnv) run(args ...string) error {
	e.t.Helper()
	isolateRoot(e.t)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// seedCache records a fresh check that found latest.
func (e *updateEnv) seedCache(latest string) {
	e.t.Helper()
	data := fmt.Sprintf(`{"last_checked":%q,"latest_version":%q}`, time.Now().UTC().Format(time.RFC3339), latest)
	if err := os.WriteFile(filepath.Join(e.configDir, "update-check.json"), []byte(data), 0o600); err != nil {
		e.t.Fatal(err)
	}
}

// useGoBin makes the running test binary look like a make install build.
func (e *updateEnv) useGoBin() {
	e.t.Helper()
	exe, err := update.ExecutablePath()
	if err != nil {
		e.t.Fatal(err)
	}
	e.t.Setenv("GOBIN", filepath.Dir(exe))
}

func TestUpdateNotifierSuppressed(t *testing.T) {
	configGet := []string{"config", "get", "server"}
	cases := []struct {
		name    string
		env     map[string]string
		version string
		noTTY   bool
		args    []string
		lookups int32 // update --check queries GitHub itself
	}{
		{name: "stderr not a terminal", noTTY: true, args: configGet},
		{name: "CI", env: map[string]string{"CI": "true"}, args: configGet},
		{name: "CUBEAPM_NO_UPDATE_NOTIFIER", env: map[string]string{"CUBEAPM_NO_UPDATE_NOTIFIER": "1"}, args: configGet},
		{name: "NO_UPDATE_NOTIFIER", env: map[string]string{"NO_UPDATE_NOTIFIER": "1"}, args: configGet},
		{name: "quiet flag", args: append([]string{"--quiet"}, configGet...)},
		{name: "quiet env", env: map[string]string{"CUBEAPM_QUIET": "1"}, args: configGet},
		{name: "dev build", version: "dev", args: configGet},
		{name: "empty version", version: " ", args: configGet},
		{name: "commit hash version", version: "0e8ddd8", args: configGet},
		{name: "version command", args: []string{"version"}},
		{name: "update command", args: []string{"update", "--check"}, lookups: 1},
		{name: "completion command", args: []string{"completion", "bash"}},
		{name: "help command", args: []string{"help"}},
		{name: "shell completion request", args: []string{"__complete", "traces", ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := newUpdateEnv(t, "v0.2.10", nil)
			if tc.version != "" {
				Version = strings.TrimSpace(tc.version)
			}
			if tc.noTTY {
				stderrIsTerminal = func() bool { return false }
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			if tc.name == "update command" {
				env.seedCache("0.2.9") // nothing new, so update --check prints no notice either
			}

			if err := env.run(tc.args...); err != nil {
				t.Fatal(err)
			}
			if pendingUpdateCheck != nil {
				t.Fatal("the update notifier started a check")
			}
			if got := env.lookups.Load(); got != tc.lookups {
				t.Fatalf("release lookups = %d, want %d", got, tc.lookups)
			}
			if env.notice.Len() != 0 {
				t.Fatalf("unexpected notice: %q", env.notice.String())
			}
		})
	}
}

func TestUpdateNotifierChecksInTerminal(t *testing.T) {
	env := newUpdateEnv(t, "v0.2.10", nil)
	if err := env.run("config", "get", "server"); err != nil {
		t.Fatal(err)
	}
	if pendingUpdateCheck == nil {
		t.Fatal("no update check started")
	}
	select {
	case <-pendingUpdateCheck.done:
	case <-time.After(5 * time.Second):
		t.Fatal("update check did not finish")
	}
	if env.lookups.Load() != 1 {
		t.Fatalf("release lookups = %d, want 1", env.lookups.Load())
	}
	if _, err := os.Stat(filepath.Join(env.configDir, "update-check.json")); err != nil {
		t.Fatalf("check result not cached: %v", err)
	}
}

func TestUpdateNoticeOncePerVersion(t *testing.T) {
	env := newUpdateEnv(t, "v0.2.10", nil)
	env.seedCache("0.2.10")
	for range 3 {
		if err := env.run("config", "get", "server"); err != nil {
			t.Fatal(err)
		}
	}
	want := "\nA new version of cubeapm is available: v0.2.9 -> v0.2.10\n" +
		"Update with: cubeapm update\n" +
		"Release notes: https://github.com/piyush-gambhir/cubeapm-cli/releases/tag/v0.2.10\n"
	if env.notice.String() != want {
		t.Fatalf("notice output:\n%q\nwant exactly one notice:\n%q", env.notice.String(), want)
	}
	if env.lookups.Load() != 0 {
		t.Fatalf("a fresh cache still reached GitHub %d times", env.lookups.Load())
	}
}

func TestUpdateNoticeGoInstallLine(t *testing.T) {
	env := newUpdateEnv(t, "v0.2.10", nil)
	env.useGoBin()
	env.seedCache("0.2.10")
	if err := env.run("config", "get", "server"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(env.notice.String(), "Update with: "+update.SourceUpdateCommand+"\n") {
		t.Fatalf("notice lacks the source-build command: %q", env.notice.String())
	}
}

func TestUpdateNoticeNeverWaitsForGitHub(t *testing.T) {
	block := make(chan struct{})
	env := newUpdateEnv(t, "v0.2.10", block)

	start := time.Now()
	err := env.run("config", "get", "server")
	elapsed := time.Since(start)
	check := pendingUpdateCheck
	// Let the background request finish before the temp dirs go away.
	t.Cleanup(func() {
		close(block)
		if check != nil {
			<-check.done
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if elapsed > time.Second {
		t.Fatalf("command waited %v for the release check", elapsed)
	}
	if env.notice.Len() != 0 {
		t.Fatalf("unexpected notice: %q", env.notice.String())
	}
}

func TestUpdateCheckCountsWhenCommandExitsFirst(t *testing.T) {
	block := make(chan struct{})
	env := newUpdateEnv(t, "v0.2.10", block)
	if err := env.run("config", "get", "server"); err != nil {
		t.Fatal(err)
	}
	first := pendingUpdateCheck
	var second *updateCheck
	// Let the background requests finish before the temp dirs go away.
	t.Cleanup(func() {
		close(block)
		for _, check := range []*updateCheck{first, second} {
			if check != nil {
				<-check.done
			}
		}
	})
	if first == nil {
		t.Fatal("no update check started")
	}

	// The first command finished while GitHub had not answered. The next one
	// must answer from the recorded attempt instead of asking again.
	if err := env.run("config", "get", "server"); err != nil {
		t.Fatal(err)
	}
	second = pendingUpdateCheck
	select {
	case <-second.done:
	default:
		t.Fatal("the second command started another GitHub request")
	}
}

func TestUpdateCheckJSON(t *testing.T) {
	env := newUpdateEnv(t, "v0.2.10", nil)
	env.seedCache("0.2.9") // --check must bypass the cache
	if err := env.run("update", "--check", "-o", "json"); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(env.out.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, env.out.String())
	}
	want := map[string]any{
		"current_version":  "0.2.9",
		"latest_version":   "0.2.10",
		"update_available": true,
		"release_url":      "https://github.com/piyush-gambhir/cubeapm-cli/releases/tag/v0.2.10",
		"install_method":   "self",
	}
	if len(got) != len(want) {
		t.Fatalf("fields = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %v, want %v", k, got[k], v)
		}
	}
	if env.lookups.Load() != 1 {
		t.Fatalf("release lookups = %d, want 1", env.lookups.Load())
	}
}

func TestUpdateCheckRunsInReadOnlyMode(t *testing.T) {
	env := newUpdateEnv(t, "v0.2.9", nil)
	if err := env.run("--read-only", "update", "--check"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(env.out.String(), "Update available: no") {
		t.Fatalf("unexpected output: %q", env.out.String())
	}
	if err := env.run("--read-only", "update", "--yes"); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("update should be blocked in read-only mode, got %v", err)
	}
}

func TestUpdateAlreadyLatest(t *testing.T) {
	env := newUpdateEnv(t, "v0.2.9", nil)
	if err := env.run("update"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(env.out.String(), "already the latest version") {
		t.Fatalf("unexpected output: %q", env.out.String())
	}
	if env.downloadHits.Load() != 0 {
		t.Fatal("downloaded a release while already up to date")
	}
}

func TestUpdateRequiresYesWithoutPrompt(t *testing.T) {
	for _, tc := range []struct {
		name  string
		args  []string
		noTTY bool
	}{
		{name: "no-input", args: []string{"--no-input", "update"}},
		{name: "stdin not a terminal", args: []string{"update"}, noTTY: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := newUpdateEnv(t, "v0.2.10", nil)
			if tc.noTTY {
				stdinIsTerminal = func() bool { return false }
			}
			err := env.run(tc.args...)
			if err == nil || !strings.Contains(err.Error(), "--yes") {
				t.Fatalf("expected an error asking for --yes, got %v", err)
			}
			if env.downloadHits.Load() != 0 {
				t.Fatal("downloaded a release without confirmation")
			}
		})
	}
}

func TestUpdateFromGoBinPrintsSourceCommand(t *testing.T) {
	env := newUpdateEnv(t, "v0.2.10", nil)
	env.useGoBin()
	if err := env.run("update", "--yes"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(env.out.String(), update.SourceUpdateCommand) {
		t.Fatalf("output lacks the source-build command: %q", env.out.String())
	}
	if env.downloadHits.Load() != 0 {
		t.Fatal("replaced a binary in a Go bin directory")
	}

	env.out.Reset()
	if err := env.run("update", "--check", "-o", "json"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(env.out.String(), `"install_method": "go"`) {
		t.Fatalf("install_method is not go: %s", env.out.String())
	}
}

func TestVersionShowsCachedLatest(t *testing.T) {
	env := newUpdateEnv(t, "v0.2.10", nil)
	if err := env.run("version"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(env.out.String(), "latest") {
		t.Fatalf("version reported a latest release without a cache: %q", env.out.String())
	}

	env.out.Reset()
	env.seedCache("0.2.10")
	if err := env.run("version"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"cubeapm version 0.2.9\n", "  latest:     0.2.10\n", "  update_available: true\n"} {
		if !strings.Contains(env.out.String(), want) {
			t.Errorf("version output lacks %q:\n%s", want, env.out.String())
		}
	}
	if env.lookups.Load() != 0 {
		t.Fatal("version used the network")
	}
}

func TestConfirmUpdate(t *testing.T) {
	for input, want := range map[string]bool{"\n": true, "y\n": true, "YES\n": true, "n\n": false, "no\n": false, "": false} {
		var out bytes.Buffer
		if got := confirmUpdate(strings.NewReader(input), &out); got != want {
			t.Errorf("confirmUpdate(%q) = %v, want %v", input, got, want)
		}
		if !strings.HasPrefix(out.String(), "Update now? [Y/n] ") {
			t.Errorf("prompt = %q", out.String())
		}
	}
}
