package update

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	cacheDuration = 24 * time.Hour
	cacheFileName = "update-check.json"
)

// UpdateInfo holds the result of an update check.
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	PublishedAt    string
}

// cacheEntry represents the cached update check result on disk.
type cacheEntry struct {
	LastChecked   time.Time `json:"last_checked"`
	LatestVersion string    `json:"latest_version"`
	ReleaseURL    string    `json:"release_url"`
}

// githubRelease represents the relevant fields from the GitHub releases API.
type githubRelease struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

// CheckForUpdate checks GitHub for a newer release. Uses a 24h cache stored
// in configDir to avoid repeated network calls.
func CheckForUpdate(currentVersion, repo, configDir string) (*UpdateInfo, error) {
	// Skip check for dev builds or empty versions.
	if currentVersion == "" || currentVersion == "dev" {
		return &UpdateInfo{CurrentVersion: currentVersion}, nil
	}

	// Try loading from cache first.
	cached, err := loadCache(configDir)
	if err == nil && time.Since(cached.LastChecked) < cacheDuration {
		return buildUpdateInfo(currentVersion, cached.LatestVersion, cached.ReleaseURL, ""), nil
	}

	return checkFresh(currentVersion, repo, configDir)
}

// CheckForUpdateFresh always performs a network call, bypassing the cache.
func CheckForUpdateFresh(currentVersion, repo, configDir string) (*UpdateInfo, error) {
	if currentVersion == "" || currentVersion == "dev" {
		return &UpdateInfo{CurrentVersion: currentVersion}, nil
	}
	return checkFresh(currentVersion, repo, configDir)
}

func checkFresh(currentVersion, repo, configDir string) (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "cubeapm-cli")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checking for update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing release response: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")

	// Update the cache.
	_ = saveCache(configDir, cacheEntry{
		LastChecked:   time.Now().UTC(),
		LatestVersion: latestVersion,
		ReleaseURL:    release.HTMLURL,
	})

	return buildUpdateInfo(currentVersion, latestVersion, release.HTMLURL, release.PublishedAt), nil
}

func buildUpdateInfo(currentVersion, latestVersion, releaseURL, publishedAt string) *UpdateInfo {
	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		ReleaseURL:     releaseURL,
		PublishedAt:    publishedAt,
	}

	current := parseSemver(currentVersion)
	latest := parseSemver(latestVersion)
	if current != nil && latest != nil {
		info.Available = compareSemver(latest, current) > 0
	}

	return info
}

// PrintUpdateNotice prints a colored notice to w if an update is available.
func PrintUpdateNotice(w io.Writer, info *UpdateInfo) {
	if info == nil || !info.Available {
		return
	}
	fmt.Fprintf(w, "\nA new version of cubeapm is available: v%s -> v%s\n", info.CurrentVersion, info.LatestVersion)
	fmt.Fprintf(w, "Run `cubeapm update` to update, or download from:\n")
	fmt.Fprintf(w, "%s\n", info.ReleaseURL)
}

// --- Semver parsing and comparison ---

type semver struct {
	Major int
	Minor int
	Patch int
}

func parseSemver(v string) *semver {
	v = strings.TrimPrefix(v, "v")
	// Strip any pre-release/metadata suffix for comparison.
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	var s semver
	n, _ := fmt.Sscanf(v, "%d.%d.%d", &s.Major, &s.Minor, &s.Patch)
	if n < 2 {
		return nil
	}
	return &s
}

// compareSemver returns >0 if a > b, 0 if equal, <0 if a < b.
func compareSemver(a, b *semver) int {
	if a.Major != b.Major {
		return a.Major - b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor - b.Minor
	}
	return a.Patch - b.Patch
}

// --- Cache helpers ---

func cachePath(configDir string) string {
	return filepath.Join(configDir, cacheFileName)
}

func loadCache(configDir string) (*cacheEntry, error) {
	data, err := os.ReadFile(cachePath(configDir))
	if err != nil {
		return nil, err
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func saveCache(configDir string, entry cacheEntry) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(configDir), data, 0600)
}

// --- Self-update functionality ---

// SelfUpdate downloads and installs the specified version of the binary,
// replacing the current executable in-place.
func SelfUpdate(version, repo string) error {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Build download URL matching the release artifact naming convention.
	archive := fmt.Sprintf("cubeapm-cli_%s_%s.tar.gz", osName, archName)
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s", repo, version, archive)

	fmt.Printf("Downloading cubeapm v%s (%s/%s)...\n", version, osName, archName)

	// Download to a temp directory.
	tmpDir, err := os.MkdirTemp("", "cubeapm-update-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archive)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		return fmt.Errorf("downloading release: %w", err)
	}

	// Extract the binary from the tarball.
	fmt.Println("Extracting...")
	binaryPath, err := extractBinary(archivePath, tmpDir, "cubeapm")
	if err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}

	// Find current executable path.
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}

	// Replace the binary atomically: copy to temp file beside the target,
	// then rename (which is atomic on the same filesystem).
	fmt.Printf("Replacing %s...\n", execPath)
	if err := replaceBinary(binaryPath, execPath); err != nil {
		return err
	}

	fmt.Printf("Successfully updated to cubeapm v%s\n", version)
	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractBinary(archivePath, destDir, binaryName string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("reading gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("reading tar: %w", err)
		}

		// Look for the binary file (may be at root or in a subdirectory).
		name := filepath.Base(header.Name)
		if name == binaryName && header.Typeflag == tar.TypeReg {
			outPath := filepath.Join(destDir, binaryName)
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return outPath, nil
		}
	}

	return "", fmt.Errorf("binary %q not found in archive", binaryName)
}

func replaceBinary(newBinary, target string) error {
	// Read the new binary into memory-ish (via temp file in same dir).
	targetDir := filepath.Dir(target)
	tmpFile, err := os.CreateTemp(targetDir, ".cubeapm-update-*")
	if err != nil {
		// If we can't write to the target directory, we may need elevated permissions.
		return fmt.Errorf("cannot write to %s (you may need to use sudo): %w", targetDir, err)
	}
	tmpPath := tmpFile.Name()

	// Copy new binary to temp file.
	src, err := os.Open(newBinary)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}

	if _, err := io.Copy(tmpFile, src); err != nil {
		src.Close()
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	src.Close()
	tmpFile.Close()

	// Preserve permissions from the original binary.
	info, err := os.Stat(target)
	if err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, info.Mode()); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Atomic rename.
	if err := os.Rename(tmpPath, target); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing binary (you may need to use sudo): %w", err)
	}

	return nil
}

// ConfirmPrompt asks the user for y/n confirmation.
func ConfirmPrompt(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}
