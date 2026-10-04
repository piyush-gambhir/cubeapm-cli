package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

type archiveEntry struct {
	name     string
	body     string
	typeflag byte        // tar only; 0 means a regular file
	mode     fs.FileMode // zip only; 0 means a regular file
}

func tarGz(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		typeflag := e.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		hdr := &tar.Header{Name: e.name, Mode: 0o755, Typeflag: typeflag}
		switch typeflag {
		case tar.TypeReg:
			hdr.Size = int64(len(e.body))
		case tar.TypeSymlink, tar.TypeLink:
			hdr.Linkname = "/etc/passwd"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func zipArchive(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		hdr := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		mode := e.mode
		if mode == 0 {
			mode = 0o755
		}
		hdr.SetMode(mode)
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(e.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// releaseServer serves one release's archive and checksums.txt and counts
// download requests.
func releaseServer(t *testing.T, version, archiveName string, archive []byte, checksums string) *atomic.Int32 {
	t.Helper()
	var hits atomic.Int32
	prefix := "/" + testRepo + "/releases/download/v" + version + "/"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case prefix + archiveName:
			w.Write(archive)
		case prefix + "checksums.txt":
			io.WriteString(w, checksums)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	orig := DownloadBaseURL
	DownloadBaseURL = srv.URL
	t.Cleanup(func() { DownloadBaseURL = orig })
	return &hits
}

func writeExecutable(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s = %q, want %q", path, got, want)
	}
}

func assertOnlyFiles(t *testing.T, dir string, want ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range entries {
		got = append(got, e.Name())
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("%s holds %v, want %v", dir, got, want)
	}
}

func TestArchiveName(t *testing.T) {
	for _, tc := range []struct{ os, arch, want string }{
		{"darwin", "arm64", "cubeapm-cli_darwin_arm64.tar.gz"},
		{"linux", "amd64", "cubeapm-cli_linux_amd64.tar.gz"},
		{"windows", "amd64", "cubeapm-cli_windows_amd64.zip"},
	} {
		if got := ArchiveName(tc.os, tc.arch); got != tc.want {
			t.Errorf("ArchiveName(%s, %s) = %s, want %s", tc.os, tc.arch, got, tc.want)
		}
	}
}

func TestInstallReplacesUnixExecutable(t *testing.T) {
	setGOOS(t, "linux", "amd64")
	archive := tarGz(t, archiveEntry{name: "README.md", body: "docs"}, archiveEntry{name: "cubeapm", body: "new binary"})
	name := "cubeapm-cli_linux_amd64.tar.gz"
	releaseServer(t, "0.2.10", name, archive, sha256Hex(archive)+"  "+name+"\n")

	dir := t.TempDir()
	target := filepath.Join(dir, "cubeapm")
	writeExecutable(t, target, "old binary")
	if err := os.Chmod(target, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := Install(context.Background(), testRepo, "0.2.10", target, io.Discard); err != nil {
		t.Fatal(err)
	}
	assertFile(t, target, "new binary")
	assertOnlyFiles(t, dir, "cubeapm")
	if runtime.GOOS != "windows" {
		info, err := os.Stat(target)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Fatalf("mode = %v, want 0755", info.Mode().Perm())
		}
	}
}

func TestInstallWindowsRenamesRunningExecutableAside(t *testing.T) {
	setGOOS(t, "windows", "amd64")
	archive := zipArchive(t, archiveEntry{name: "LICENSE", body: "license"}, archiveEntry{name: "cubeapm.exe", body: "new exe"})
	name := "cubeapm-cli_windows_amd64.zip"
	releaseServer(t, "0.2.10", name, archive, sha256Hex(archive)+"  "+name+"\n")

	dir := t.TempDir()
	target := filepath.Join(dir, "cubeapm.exe")
	writeExecutable(t, target, "old exe")
	writeExecutable(t, target+".old", "older exe from a previous update")

	if err := Install(context.Background(), testRepo, "0.2.10", target, io.Discard); err != nil {
		t.Fatal(err)
	}
	assertFile(t, target, "new exe")
	assertFile(t, target+".old", "old exe")
	assertOnlyFiles(t, dir, "cubeapm.exe", "cubeapm.exe.old")

	// The next start removes the leftover.
	removeOldExecutable("windows", func() (string, error) { return target, nil })
	assertOnlyFiles(t, dir, "cubeapm.exe")
}

func TestRemoveOldExecutableOnlyOnWindows(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cubeapm")
	writeExecutable(t, target+".old", "keep")
	removeOldExecutable("linux", func() (string, error) { return target, nil })
	assertFile(t, target+".old", "keep")
}

func TestReplaceExecutableWindowsRestoresOnFailure(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cubeapm.exe")
	writeExecutable(t, target, "old exe")
	newBinary := filepath.Join(t.TempDir(), "cubeapm.exe")
	writeExecutable(t, newBinary, "new exe")

	calls := 0
	rename := func(oldpath, newpath string) error {
		calls++
		if calls == 2 { // moving the new binary into place
			return errors.New("simulated failure")
		}
		return os.Rename(oldpath, newpath)
	}
	if err := replaceExecutable("windows", newBinary, target, rename); err == nil {
		t.Fatal("expected an error")
	}
	assertFile(t, target, "old exe")
	assertOnlyFiles(t, dir, "cubeapm.exe")
}

func TestInstallRefusesBadChecksums(t *testing.T) {
	setGOOS(t, "linux", "amd64")
	archive := tarGz(t, archiveEntry{name: "cubeapm", body: "tampered binary"})
	name := "cubeapm-cli_linux_amd64.tar.gz"
	for label, checksums := range map[string]string{
		"mismatch":      sha256Hex([]byte("something else")) + "  " + name + "\n",
		"missing entry": sha256Hex(archive) + "  cubeapm-cli_darwin_arm64.tar.gz\n",
		"empty":         "",
	} {
		t.Run(label, func(t *testing.T) {
			releaseServer(t, "0.2.10", name, archive, checksums)
			dir := t.TempDir()
			target := filepath.Join(dir, "cubeapm")
			writeExecutable(t, target, "old binary")

			err := Install(context.Background(), testRepo, "0.2.10", target, io.Discard)
			if err == nil || !strings.Contains(err.Error(), "checksum") {
				t.Fatalf("expected a checksum error, got %v", err)
			}
			assertFile(t, target, "old binary")
			assertOnlyFiles(t, dir, "cubeapm")
		})
	}
}

func TestExtractRefusesUnsafeEntries(t *testing.T) {
	tarCases := map[string][]archiveEntry{
		"parent traversal":     {{name: "../cubeapm", body: "x"}},
		"nested traversal":     {{name: "dist/../../cubeapm", body: "x"}},
		"absolute path":        {{name: "/cubeapm", body: "x"}},
		"traversal before bin": {{name: "../../etc/cron.d/x", body: "x"}, {name: "cubeapm", body: "x"}},
		"symlink binary":       {{name: "cubeapm", typeflag: tar.TypeSymlink}},
		"hard link binary":     {{name: "cubeapm", typeflag: tar.TypeLink}},
		"directory binary":     {{name: "cubeapm", typeflag: tar.TypeDir}},
		"binary not at root":   {{name: "nested/cubeapm", body: "x"}},
	}
	for label, entries := range tarCases {
		t.Run("tar/"+label, func(t *testing.T) {
			dir := t.TempDir()
			archivePath := filepath.Join(dir, "a.tar.gz")
			if err := os.WriteFile(archivePath, tarGz(t, entries...), 0o600); err != nil {
				t.Fatal(err)
			}
			out := t.TempDir()
			if _, err := extractBinary(archivePath, out, "cubeapm"); err == nil {
				t.Fatal("expected extraction to fail")
			}
			assertOnlyFiles(t, out)
		})
	}

	zipCases := map[string][]archiveEntry{
		"parent traversal": {{name: "../cubeapm.exe", body: "x"}},
		"backslash escape": {{name: `..\cubeapm.exe`, body: "x"}},
		"drive letter":     {{name: "C:/cubeapm.exe", body: "x"}},
		"symlink binary":   {{name: "cubeapm.exe", body: "C:/Windows/notepad.exe", mode: fs.ModeSymlink | 0o777}},
	}
	for label, entries := range zipCases {
		t.Run("zip/"+label, func(t *testing.T) {
			dir := t.TempDir()
			archivePath := filepath.Join(dir, "a.zip")
			if err := os.WriteFile(archivePath, zipArchive(t, entries...), 0o600); err != nil {
				t.Fatal(err)
			}
			out := t.TempDir()
			if _, err := extractBinary(archivePath, out, "cubeapm.exe"); err == nil {
				t.Fatal("expected extraction to fail")
			}
			assertOnlyFiles(t, out)
		})
	}
}

func TestExtractAcceptsDotSlashEntry(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "a.tar.gz")
	if err := os.WriteFile(archivePath, tarGz(t, archiveEntry{name: "./cubeapm", body: "bin"}), 0o600); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	got, err := extractBinary(archivePath, out, "cubeapm")
	if err != nil {
		t.Fatal(err)
	}
	assertFile(t, got, "bin")
}

func TestInstallUnwritableDirectoryKeepsBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permissions differ on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root can write to read-only directories")
	}
	setGOOS(t, "linux", "amd64")
	archive := tarGz(t, archiveEntry{name: "cubeapm", body: "new binary"})
	name := "cubeapm-cli_linux_amd64.tar.gz"
	hits := releaseServer(t, "0.2.10", name, archive, sha256Hex(archive)+"  "+name+"\n")

	dir := t.TempDir()
	target := filepath.Join(dir, "cubeapm")
	writeExecutable(t, target, "old binary")
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })

	err := Install(context.Background(), testRepo, "0.2.10", target, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "sudo") || !strings.Contains(err.Error(), "install script") {
		t.Fatalf("expected a not-writable error suggesting sudo or the install script, got %v", err)
	}
	assertFile(t, target, "old binary")
	if hits.Load() != 0 {
		t.Fatalf("downloaded %d files before noticing the directory is not writable", hits.Load())
	}
}

func TestInstallRejectsInvalidVersion(t *testing.T) {
	target := filepath.Join(t.TempDir(), "cubeapm")
	writeExecutable(t, target, "old")
	if err := Install(context.Background(), testRepo, "../0.2.10", target, io.Discard); err == nil {
		t.Fatal("expected an error")
	}
	assertFile(t, target, "old")
}

func TestCopyLimited(t *testing.T) {
	var buf bytes.Buffer
	if err := copyLimited(&buf, strings.NewReader("12345"), 5); err != nil {
		t.Fatalf("at the limit: %v", err)
	}
	if err := copyLimited(io.Discard, strings.NewReader("123456"), 5); err == nil {
		t.Fatal("expected an error past the limit")
	}
}
