package update

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

const testRepo = "piyush-gambhir/cubeapm-cli"

func setGOOS(t *testing.T, osName, arch string) {
	t.Helper()
	origOS, origArch := goos, goarch
	goos, goarch = osName, arch
	t.Cleanup(func() { goos, goarch = origOS, origArch })
}

// latestServer answers releases/latest like github.com: a 302 to the tag page
// (or status when non-zero). It counts lookups and requests to the tag page,
// which a correct client never makes.
func latestServer(t *testing.T, tag string, status int) (lookups, followed *atomic.Int32) {
	t.Helper()
	return redirectServer(t, status, func(base string) string {
		return base + "/" + testRepo + "/releases/tag/" + tag
	})
}

// redirectServer serves releases/latest with a 302 to location(serverURL), or
// with status when non-zero. An empty location sends no Location header.
func redirectServer(t *testing.T, status int, location func(base string) string) (lookups, followed *atomic.Int32) {
	t.Helper()
	lookups, followed = new(atomic.Int32), new(atomic.Int32)
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+testRepo+"/releases/latest" {
			followed.Add(1)
			w.WriteHeader(http.StatusOK)
			return
		}
		lookups.Add(1)
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		if loc := location(srv.URL); loc != "" {
			w.Header().Set("Location", loc)
		}
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)
	orig := GitHubBaseURL
	GitHubBaseURL = srv.URL
	t.Cleanup(func() { GitHubBaseURL = orig })
	return lookups, followed
}

func TestIsReleaseVersion(t *testing.T) {
	for v, want := range map[string]bool{
		"0.2.9": true, "v0.2.9": true, "v0.2.9-3-gabc1234-dirty": true, "1.0.0+build.1": true,
		"dev": false, "": false, "0e8ddd8": false, "0.2": false, "v1.2.3/../x": false,
	} {
		if got := IsReleaseVersion(v); got != want {
			t.Errorf("IsReleaseVersion(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestFetchLatestCachesSuccessAndFailure(t *testing.T) {
	dir := t.TempDir()
	latestServer(t, "v0.2.10", 0)

	info, err := FetchLatest(context.Background(), "0.2.9", testRepo, dir, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Available || info.LatestVersion != "0.2.10" || info.ReleaseURL != "https://github.com/"+testRepo+"/releases/tag/v0.2.10" {
		t.Fatalf("unexpected info %+v", info)
	}
	if cached, ok := CachedCheck("0.2.9", testRepo, dir, time.Now()); !ok || cached == nil || !cached.Available {
		t.Fatalf("CachedCheck after success = %+v, %v", cached, ok)
	}
	if _, ok := CachedCheck("0.2.9", testRepo, dir, time.Now().Add(25*time.Hour)); ok {
		t.Fatal("cache older than 24h should be stale")
	}

	// A failed lookup is cached as well, so the next command does not retry;
	// the last known release survives the failure.
	hits, _ := latestServer(t, "", http.StatusForbidden)
	if _, err := FetchLatest(context.Background(), "0.2.9", testRepo, dir, time.Second); err == nil {
		t.Fatal("expected an error from a failing GitHub")
	}
	cached, ok := CachedCheck("0.2.9", testRepo, dir, time.Now())
	if !ok {
		t.Fatal("a failed check should be cached")
	}
	if cached == nil || cached.LatestVersion != "0.2.10" {
		t.Fatalf("failed check lost the last known release: %+v", cached)
	}
	if hits.Load() != 1 {
		t.Fatalf("GitHub hits = %d, want 1", hits.Load())
	}
}

func TestFetchLatestReadsTagFromRedirect(t *testing.T) {
	lookups, followed := latestServer(t, "v0.2.10", 0)
	info, err := FetchLatest(context.Background(), "0.2.9", testRepo, t.TempDir(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if info.LatestVersion != "0.2.10" || info.ReleaseURL != "https://github.com/"+testRepo+"/releases/tag/v0.2.10" {
		t.Fatalf("unexpected info %+v", info)
	}
	if lookups.Load() != 1 || followed.Load() != 0 {
		t.Fatalf("lookups = %d, followed redirects = %d; want 1 and 0", lookups.Load(), followed.Load())
	}
}

func TestFetchLatestRejectsBadRedirects(t *testing.T) {
	for name, tc := range map[string]struct {
		status   int
		location func(base string) string
	}{
		"missing Location": {location: func(string) string { return "" }},
		"foreign host": {location: func(string) string {
			return "https://evil.example.com/" + testRepo + "/releases/tag/v0.2.10"
		}},
		"other repo": {location: func(base string) string { return base + "/someone/else/releases/tag/v0.2.10" }},
		"non-semver tag": {location: func(base string) string {
			return base + "/" + testRepo + "/releases/tag/nightly"
		}},
		"tag without v": {location: func(base string) string { return base + "/" + testRepo + "/releases/tag/0.2.10" }},
		"path traversal": {location: func(base string) string {
			return base + "/" + testRepo + "/releases/tag/v0.2.10/../../evil"
		}},
		"no release (releases page)": {location: func(base string) string { return base + "/" + testRepo + "/releases" }},
		"not a redirect":             {status: http.StatusOK},
		"rate limited":               {status: http.StatusTooManyRequests},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			_, followed := redirectServer(t, tc.status, tc.location)
			if _, err := FetchLatest(context.Background(), "0.2.9", testRepo, dir, time.Second); err == nil {
				t.Fatal("expected an error")
			}
			if followed.Load() != 0 {
				t.Fatal("the client followed the redirect")
			}
			// The background check caches the failure.
			if info, ok := CachedCheck("0.2.9", testRepo, dir, time.Now()); !ok || info != nil {
				t.Fatalf("CachedCheck after a bad redirect = %+v, %v; want a cached failure", info, ok)
			}
		})
	}
}

func TestRecordCheckAttempt(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	notifiedAt := now.Add(-time.Hour).UTC().Truncate(time.Second)
	if err := saveCache(dir, cacheEntry{LastChecked: now.Add(-48 * time.Hour), LatestVersion: "0.2.10",
		NotifiedVersion: "0.2.10", NotifiedAt: notifiedAt}); err != nil {
		t.Fatal(err)
	}
	if _, ok := CachedCheck("0.2.9", testRepo, dir, now); ok {
		t.Fatal("a 48h-old check should be stale")
	}
	if err := RecordCheckAttempt(dir, now); err != nil {
		t.Fatal(err)
	}
	// Until the answer arrives the attempt counts as a check that kept the last
	// known release, so the next command does not ask GitHub again.
	info, ok := CachedCheck("0.2.9", testRepo, dir, now.Add(time.Minute))
	if !ok || info == nil || info.LatestVersion != "0.2.10" {
		t.Fatalf("CachedCheck after an attempt = %+v, %v", info, ok)
	}
	cached, err := loadCache(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cached.NotifiedVersion != "0.2.10" || !cached.NotifiedAt.Equal(notifiedAt) {
		t.Fatalf("attempt dropped the notice record: %+v", cached)
	}
}

func TestFetchLatestSkipsDevBuilds(t *testing.T) {
	hits, _ := latestServer(t, "v0.2.10", 0)
	for _, v := range []string{"dev", "", "0e8ddd8"} {
		if _, err := FetchLatest(context.Background(), v, testRepo, t.TempDir(), time.Second); err == nil {
			t.Errorf("version %q: expected an error", v)
		}
	}
	if hits.Load() != 0 {
		t.Fatalf("dev builds reached GitHub %d times", hits.Load())
	}
}

func TestNotifyOncePerVersionPerDay(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	info := buildUpdateInfo("0.2.9", "0.2.10", testRepo)

	var buf bytes.Buffer
	if !Notify(&buf, dir, info, SelfUpdateCommand, now) {
		t.Fatal("first notice was not printed")
	}
	want := "\nA new version of cubeapm is available: v0.2.9 -> v0.2.10\n" +
		"Update with: cubeapm update\n" +
		"Release notes: https://github.com/piyush-gambhir/cubeapm-cli/releases/tag/v0.2.10\n"
	if buf.String() != want {
		t.Fatalf("notice:\n%q\nwant:\n%q", buf.String(), want)
	}

	buf.Reset()
	if Notify(&buf, dir, info, SelfUpdateCommand, now.Add(time.Hour)) || buf.Len() != 0 {
		t.Fatalf("same release announced twice within 24h: %q", buf.String())
	}
	if !Notify(&buf, dir, buildUpdateInfo("0.2.9", "0.2.11", testRepo), SelfUpdateCommand, now.Add(2*time.Hour)) {
		t.Fatal("a newer release should be announced right away")
	}
	if !Notify(&buf, dir, buildUpdateInfo("0.2.9", "0.2.11", testRepo), SelfUpdateCommand, now.Add(27*time.Hour)) {
		t.Fatal("the release should be announced again after 24h")
	}

	var entry cacheEntry
	data, _ := os.ReadFile(filepath.Join(dir, cacheFileName))
	if err := json.Unmarshal(data, &entry); err != nil || entry.NotifiedVersion != "0.2.11" {
		t.Fatalf("cache does not record the notice: %s", data)
	}
}

func TestNotifyGoInstallLine(t *testing.T) {
	var buf bytes.Buffer
	Notify(&buf, t.TempDir(), buildUpdateInfo("v0.2.9", "0.2.10", testRepo), UpdateCommand(InstallGo), time.Now())
	if !bytes.Contains(buf.Bytes(), []byte("Update with: git pull && make install (in your cubeapm-cli/cli-go checkout)\n")) {
		t.Fatalf("go install notice lacks the source-build command: %q", buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte("v0.2.9 -> v0.2.10")) {
		t.Fatalf("notice does not normalize versions: %q", buf.String())
	}
}

func TestDetectInstallMethod(t *testing.T) {
	gobin, gopath, home, other := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("GOBIN", gobin)
	t.Setenv("GOPATH", gopath)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	for path, want := range map[string]string{
		filepath.Join(gobin, "cubeapm"):              InstallGo,
		filepath.Join(gopath, "bin", "cubeapm"):      InstallGo,
		filepath.Join(home, "go", "bin", "cubeapm"):  InstallGo,
		filepath.Join(other, "cubeapm"):              InstallSelf,
		filepath.Join(home, "go", "cubeapm", "bin"):  InstallSelf,
		filepath.Join(home, ".local", "bin", "cube"): InstallSelf,
	} {
		if got := DetectInstallMethod(path); got != want {
			t.Errorf("DetectInstallMethod(%s) = %s, want %s", path, got, want)
		}
	}
}

func TestCachedUpdateInfo(t *testing.T) {
	dir := t.TempDir()
	if CachedUpdateInfo("0.2.9", testRepo, dir) != nil {
		t.Fatal("expected nil without a cache")
	}
	if err := saveCache(dir, cacheEntry{LastChecked: time.Now().Add(-72 * time.Hour), LatestVersion: "0.2.10"}); err != nil {
		t.Fatal(err)
	}
	if info := CachedUpdateInfo("0.2.9", testRepo, dir); info == nil || !info.Available || info.LatestVersion != "0.2.10" {
		t.Fatalf("CachedUpdateInfo = %+v", info)
	}
	// After a manual upgrade the cached release is older than this binary.
	if info := CachedUpdateInfo("0.2.11", testRepo, dir); info != nil {
		t.Fatalf("stale cache reported %+v", info)
	}
	if CachedUpdateInfo("dev", testRepo, dir) != nil {
		t.Fatal("dev builds should not report a latest version")
	}
}
