package update

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

const testRepo = "piyush-gambhir/cubeapm-cli"

func setGOOS(t *testing.T, value string) {
	t.Helper()
	orig := goos
	goos = value
	t.Cleanup(func() { goos = orig })
}

func TestSelfUpdateRefusesOnWindows(t *testing.T) {
	setGOOS(t, "windows")
	// A canceled context makes any network attempt fail fast, so the test
	// only passes when the Windows check runs before downloading.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := SelfUpdateContext(ctx, "0.2.8", testRepo)
	if err == nil {
		t.Fatal("expected an error on Windows")
	}
	for _, want := range []string{"cubeapm.exe", "https://github.com/" + testRepo + "/releases/tag/v0.2.8"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestCheckInstallSupported(t *testing.T) {
	setGOOS(t, "linux")
	if err := CheckInstallSupported("0.2.8", testRepo); err != nil {
		t.Fatalf("linux: unexpected error: %v", err)
	}
	setGOOS(t, "windows")
	if err := CheckInstallSupported("0.2.8", testRepo); err == nil {
		t.Fatal("windows: expected an error")
	}
}

func TestPrintUpdateNoticeOnWindows(t *testing.T) {
	setGOOS(t, "windows")
	info := &UpdateInfo{Available: true, CurrentVersion: "0.2.7", LatestVersion: "0.2.8", ReleaseURL: "https://example.test/v0.2.8"}
	var buf bytes.Buffer
	PrintUpdateNotice(&buf, info)
	out := buf.String()
	if strings.Contains(out, "cubeapm update") {
		t.Errorf("Windows notice suggests self-update: %q", out)
	}
	if !strings.Contains(out, info.ReleaseURL) || !strings.Contains(out, "cubeapm.exe") {
		t.Errorf("Windows notice lacks release URL or cubeapm.exe: %q", out)
	}
}
