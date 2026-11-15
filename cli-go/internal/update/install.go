package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	InstallSelf = "self"
	InstallGo   = "go"
)

const (
	SelfUpdateCommand   = "cubeapm update"
	SourceUpdateCommand = "git pull && make install (in your cubeapm-cli/cli-go checkout)"
)

const maxChecksumsBytes = 1 << 20

var goarch = runtime.GOARCH

func ExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("finding current executable: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}
	return resolved, nil
}

func DetectInstallMethod(execPath string) string {
	dir := canonicalDir(filepath.Dir(execPath))
	for _, goBin := range goBinDirs() {
		if sameDir(dir, canonicalDir(goBin)) {
			return InstallGo
		}
	}
	return InstallSelf
}

func UpdateCommand(installMethod string) string {
	if installMethod == InstallGo {
		return SourceUpdateCommand
	}
	return SelfUpdateCommand
}

func goBinDirs() []string {
	var dirs []string
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		dirs = append(dirs, gobin)
	}
	for _, gopath := range filepath.SplitList(os.Getenv("GOPATH")) {
		if gopath != "" {
			dirs = append(dirs, filepath.Join(gopath, "bin"))
		}
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		dirs = append(dirs, filepath.Join(home, "go", "bin"))
	}
	return dirs
}

func canonicalDir(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	return filepath.Clean(dir)
}

func sameDir(a, b string) bool {
	if goos == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func ArchiveName(osName, arch string) string {
	ext := ".tar.gz"
	if osName == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("%s_%s_%s%s", project, osName, arch, ext)
}

func binaryFileName(osName string) string {
	if osName == "windows" {
		return binName + ".exe"
	}
	return binName
}

func Install(ctx context.Context, repo, version, execPath string, progress io.Writer) error {
	if !IsReleaseVersion(version) {
		return fmt.Errorf("invalid release version %q", version)
	}
	version = NormalizeVersion(version)
	if err := checkWritable(filepath.Dir(execPath)); err != nil {
		return err
	}

	archive := ArchiveName(goos, goarch)
	baseURL := fmt.Sprintf("%s/%s/releases/download/v%s/", GitHubBaseURL, repo, version)

	tmpDir, err := os.MkdirTemp("", "cubeapm-update-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Fprintf(progress, "Downloading %s...\n", archive)
	archivePath := filepath.Join(tmpDir, archive)
	if err := downloadFile(ctx, baseURL+archive, archivePath, maxReleaseArtifactBytes); err != nil {
		return fmt.Errorf("downloading %s: %w", archive, err)
	}
	checksumsPath := filepath.Join(tmpDir, "checksums.txt")
	if err := downloadFile(ctx, baseURL+"checksums.txt", checksumsPath, maxChecksumsBytes); err != nil {
		return fmt.Errorf("downloading checksums.txt: %w", err)
	}

	fmt.Fprintln(progress, "Verifying SHA-256 checksum...")
	if err := verifyChecksum(archivePath, checksumsPath, archive); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	newBinary, err := extractBinary(archivePath, tmpDir, binaryFileName(goos))
	if err != nil {
		return fmt.Errorf("extracting %s: %w", archive, err)
	}

	fmt.Fprintf(progress, "Replacing %s...\n", execPath)
	return replaceExecutable(goos, newBinary, execPath, os.Rename)
}

func checkWritable(dir string) error {
	probe, err := os.CreateTemp(dir, ".cubeapm-write-test-*")
	if err != nil {
		if goos == "windows" {
			return fmt.Errorf("cannot write to %s: run the terminal as Administrator, or reinstall cubeapm.exe into a directory you can write to: %w", dir, err)
		}
		return fmt.Errorf("cannot write to %s: re-run with sudo (sudo cubeapm update), or reinstall with the install script into a writable directory (INSTALL_DIR=~/.local/bin): %w", dir, err)
	}
	probe.Close()
	return os.Remove(probe.Name())
}

func downloadFile(ctx context.Context, url, dest string, limit int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", project)
	resp, err := (&http.Client{Timeout: 5 * time.Minute}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if err := copyLimited(f, resp.Body, limit); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func verifyChecksum(filePath, checksumsPath, archiveName string) error {
	data, err := os.ReadFile(checksumsPath)
	if err != nil {
		return fmt.Errorf("reading checksums file: %w", err)
	}
	var expected string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == archiveName {
			expected = strings.ToLower(fields[0])
			break
		}
	}
	if len(expected) != sha256.Size*2 {
		return fmt.Errorf("no SHA-256 checksum for %s in checksums.txt", archiveName)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if err := copyLimited(h, f, maxReleaseArtifactBytes); err != nil {
		return err
	}
	if actual := hex.EncodeToString(h.Sum(nil)); actual != expected {
		return fmt.Errorf("SHA-256 mismatch for %s: expected %s, got %s", archiveName, expected, actual)
	}
	return nil
}

func extractBinary(archivePath, destDir, binaryName string) (string, error) {
	outPath := filepath.Join(destDir, binaryName)
	if strings.HasSuffix(archivePath, ".zip") {
		return outPath, extractFromZip(archivePath, outPath, binaryName)
	}
	return outPath, extractFromTarGz(archivePath, outPath, binaryName)
}

func extractFromTarGz(archivePath, outPath, binaryName string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("reading gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("%s not found in archive", binaryName)
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}
		name, err := archiveEntryName(header.Name)
		if err != nil {
			return err
		}
		if name != binaryName {
			continue
		}
		if header.Typeflag != tar.TypeReg {
			return fmt.Errorf("archive entry %q is not a regular file", header.Name)
		}
		if header.Size > maxReleaseArtifactBytes {
			return fmt.Errorf("archive entry %q exceeds %d MiB limit", header.Name, maxReleaseArtifactBytes>>20)
		}
		return writeNewFile(outPath, tr)
	}
}

func extractFromZip(archivePath, outPath, binaryName string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("reading zip: %w", err)
	}
	defer zr.Close()
	for _, entry := range zr.File {
		name, err := archiveEntryName(entry.Name)
		if err != nil {
			return err
		}
		if name != binaryName {
			continue
		}
		if !entry.Mode().IsRegular() {
			return fmt.Errorf("archive entry %q is not a regular file", entry.Name)
		}
		if entry.UncompressedSize64 > uint64(maxReleaseArtifactBytes) {
			return fmt.Errorf("archive entry %q exceeds %d MiB limit", entry.Name, maxReleaseArtifactBytes>>20)
		}
		rc, err := entry.Open()
		if err != nil {
			return err
		}
		err = writeNewFile(outPath, rc)
		rc.Close()
		return err
	}
	return fmt.Errorf("%s not found in archive", binaryName)
}

func archiveEntryName(name string) (string, error) {
	slashed := strings.ReplaceAll(name, `\`, "/")
	if slashed == "" || strings.HasPrefix(slashed, "/") || (len(slashed) >= 2 && slashed[1] == ':') {
		return "", fmt.Errorf("unsafe path %q in archive", name)
	}
	for _, part := range strings.Split(slashed, "/") {
		if part == ".." {
			return "", fmt.Errorf("unsafe path %q in archive", name)
		}
	}
	return path.Clean(slashed), nil
}

func writeNewFile(outPath string, src io.Reader) error {
	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if err := copyLimited(out, src, maxReleaseArtifactBytes); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func replaceExecutable(osName, newBinary, target string, rename func(oldpath, newpath string) error) error {
	staged, err := stageBinary(newBinary, filepath.Dir(target))
	if err != nil {
		return err
	}
	defer os.Remove(staged)

	if osName != "windows" {
		if err := rename(staged, target); err != nil {
			return fmt.Errorf("replacing %s: %w", target, err)
		}
		return nil
	}

	old := target + ".old"
	if err := os.Remove(old); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing leftover %s: %w", old, err)
	}
	if err := rename(target, old); err != nil {
		return fmt.Errorf("moving %s aside: %w", target, err)
	}
	if err := rename(staged, target); err != nil {
		if restoreErr := rename(old, target); restoreErr != nil {
			return fmt.Errorf("installing new %s: %w (restoring the old binary also failed: %v; it is at %s)", target, err, restoreErr, old)
		}
		return fmt.Errorf("installing new %s: %w", target, err)
	}
	return nil
}

func stageBinary(newBinary, dir string) (string, error) {
	src, err := os.Open(newBinary)
	if err != nil {
		return "", err
	}
	defer src.Close()

	tmp, err := os.CreateTemp(dir, ".cubeapm-update-*")
	if err != nil {
		return "", fmt.Errorf("cannot write to %s: %w", dir, err)
	}
	if err := copyLimited(tmp, src, maxReleaseArtifactBytes); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

func RemoveOldExecutable() {
	removeOldExecutable(goos, ExecutablePath)
}

func removeOldExecutable(osName string, executable func() (string, error)) {
	if osName != "windows" {
		return
	}
	exe, err := executable()
	if err != nil {
		return
	}
	_ = os.Remove(exe + ".old")
}
