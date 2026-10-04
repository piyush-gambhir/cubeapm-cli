// Package update checks GitHub for new cubeapm releases, prints the update
// notice, and installs releases in place.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	cacheDuration  = 24 * time.Hour
	noticeInterval = 24 * time.Hour
	cacheFileName  = "update-check.json"

	// BackgroundTimeout bounds the release lookup behind the update notice.
	BackgroundTimeout = 3 * time.Second
	// CommandTimeout bounds the release lookup of `cubeapm update`.
	CommandTimeout = 15 * time.Second

	binName = "cubeapm"
	project = "cubeapm-cli"
)

// GitHubBaseURL serves the releases/latest redirect and the release assets.
// Neither goes through api.github.com, whose unauthenticated limit of 60
// requests per hour per IP breaks shared networks. Tests point this at an
// httptest server; release notes links always use github.com.
var GitHubBaseURL = "https://github.com"

// goos is the OS the binary runs on. Tests override it to cover Windows.
var goos = runtime.GOOS

// UpdateInfo holds the result of an update check. Versions carry no "v".
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
}

// cacheEntry is the update-check.json file in the config directory.
type cacheEntry struct {
	LastChecked     time.Time `json:"last_checked"`
	LatestVersion   string    `json:"latest_version,omitempty"`
	CheckFailed     bool      `json:"check_failed,omitempty"`
	NotifiedVersion string    `json:"notified_version,omitempty"`
	NotifiedAt      time.Time `json:"notified_at,omitzero"`
}

var semverPattern = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

// IsReleaseVersion reports whether v is a semantic version. Development builds
// ("dev", empty, or a bare commit hash) are not, and never check for updates.
func IsReleaseVersion(v string) bool {
	return semverPattern.MatchString(v)
}

// NormalizeVersion strips a leading "v" so versions print as v<version>.
func NormalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

// ReleaseURL is the release notes page for version.
func ReleaseURL(repo, version string) string {
	return fmt.Sprintf("https://github.com/%s/releases/tag/v%s", repo, NormalizeVersion(version))
}

// CachedCheck returns the cached result when the last check is less than 24
// hours old; ok is false when GitHub should be asked again. info is nil when
// the last check failed and nothing is known. It never touches the network.
func CachedCheck(currentVersion, repo, configDir string, now time.Time) (info *UpdateInfo, ok bool) {
	cached, err := loadCache(configDir)
	if err != nil {
		return nil, false
	}
	age := now.Sub(cached.LastChecked)
	if age < 0 || age >= cacheDuration {
		return nil, false
	}
	if !IsReleaseVersion(cached.LatestVersion) {
		return nil, true
	}
	return buildUpdateInfo(currentVersion, cached.LatestVersion, repo), true
}

// FetchLatest asks GitHub for the latest release and caches the result. A
// failed lookup is cached too, so a broken network does not retry on every
// command.
func FetchLatest(ctx context.Context, currentVersion, repo, configDir string, timeout time.Duration) (*UpdateInfo, error) {
	if !IsReleaseVersion(currentVersion) {
		return nil, fmt.Errorf("development build %q does not check for updates", currentVersion)
	}
	return checkFresh(ctx, currentVersion, repo, configDir, timeout)
}

// RecordCheckAttempt records that a release check is starting, before any
// request is sent. A command that exits before the answer arrives still counts
// as a (failed) check, so GitHub is asked at most once a day. The last known
// release and the notice record are kept.
func RecordCheckAttempt(configDir string, now time.Time) error {
	var entry cacheEntry
	if cached, err := loadCache(configDir); err == nil {
		entry = *cached
	}
	entry.LastChecked = now.UTC()
	entry.CheckFailed = true // replaced once the answer arrives
	return saveCache(configDir, entry)
}

// CachedUpdateInfo reads the last known release from the cache without any
// network access. It returns nil when nothing useful is cached.
func CachedUpdateInfo(currentVersion, repo, configDir string) *UpdateInfo {
	if !IsReleaseVersion(currentVersion) {
		return nil
	}
	cached, err := loadCache(configDir)
	if err != nil || !IsReleaseVersion(cached.LatestVersion) {
		return nil
	}
	// A cached release older than this binary predates a manual upgrade.
	if compareSemver(parseSemver(cached.LatestVersion), parseSemver(currentVersion)) < 0 {
		return nil
	}
	return buildUpdateInfo(currentVersion, cached.LatestVersion, repo)
}

func checkFresh(ctx context.Context, currentVersion, repo, configDir string, timeout time.Duration) (*UpdateInfo, error) {
	latest, err := fetchLatestVersion(ctx, repo, timeout)
	cached, _ := loadCache(configDir)
	entry := cacheEntry{LastChecked: time.Now().UTC()}
	if cached != nil {
		entry.NotifiedVersion = cached.NotifiedVersion
		entry.NotifiedAt = cached.NotifiedAt
	}
	if err != nil {
		entry.CheckFailed = true
		if cached != nil {
			entry.LatestVersion = cached.LatestVersion
		}
		_ = saveCache(configDir, entry)
		return nil, err
	}
	entry.LatestVersion = latest
	_ = saveCache(configDir, entry)
	return buildUpdateInfo(currentVersion, latest, repo), nil
}

// fetchLatestVersion reads the latest release tag from the redirect GitHub
// sends for /<repo>/releases/latest (302 to /<repo>/releases/tag/<tag>). The
// redirect is never followed, so one small request answers the question.
func fetchLatestVersion(ctx context.Context, repo string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	base, err := url.Parse(GitHubBaseURL)
	if err != nil {
		return "", fmt.Errorf("parsing GitHub URL: %w", err)
	}
	latestURL := fmt.Sprintf("%s/%s/releases/latest", strings.TrimSuffix(GitHubBaseURL, "/"), repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", project)

	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("checking for update: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("%s returned status %d, expected a redirect to the latest release", latestURL, resp.StatusCode)
	}
	location, err := resp.Location()
	if err != nil {
		return "", fmt.Errorf("%s redirect has no usable Location header: %w", latestURL, err)
	}
	if location.Scheme != base.Scheme || location.Host != base.Host {
		return "", fmt.Errorf("latest release redirect points to unexpected host %q", location.Host)
	}
	prefix := "/" + repo + "/releases/tag/"
	tag := ""
	if len(location.Path) > len(prefix) && strings.EqualFold(location.Path[:len(prefix)], prefix) {
		tag = location.Path[len(prefix):]
	}
	// The tag becomes part of download URLs, so accept only a plain v<semver>.
	if !strings.HasPrefix(tag, "v") || !IsReleaseVersion(tag) {
		return "", fmt.Errorf("latest release redirect %q does not name a version tag", location.String())
	}
	return NormalizeVersion(tag), nil
}

func buildUpdateInfo(currentVersion, latestVersion, repo string) *UpdateInfo {
	info := &UpdateInfo{
		CurrentVersion: NormalizeVersion(currentVersion),
		LatestVersion:  NormalizeVersion(latestVersion),
		ReleaseURL:     ReleaseURL(repo, latestVersion),
	}
	current := parseSemver(currentVersion)
	latest := parseSemver(latestVersion)
	if current != nil && latest != nil {
		info.Available = compareSemver(latest, current) > 0
	}
	return info
}

// Notify prints the update notice to w unless the same release was already
// announced within the last 24 hours, and records the announcement.
func Notify(w io.Writer, configDir string, info *UpdateInfo, updateCommand string, now time.Time) bool {
	if info == nil || !info.Available {
		return false
	}
	cached, _ := loadCache(configDir)
	if cached == nil {
		cached = &cacheEntry{}
	}
	since := now.Sub(cached.NotifiedAt)
	if cached.NotifiedVersion == info.LatestVersion && since >= 0 && since < noticeInterval {
		return false
	}
	cached.NotifiedVersion = info.LatestVersion
	cached.NotifiedAt = now.UTC()
	if err := saveCache(configDir, *cached); err != nil {
		// Without a record the notice would repeat on every command.
		return false
	}
	PrintNotice(w, info, updateCommand)
	return true
}

// PrintNotice writes the update notice, preceded by a blank line.
func PrintNotice(w io.Writer, info *UpdateInfo, updateCommand string) {
	fmt.Fprintf(w, "\nA new version of %s is available: v%s -> v%s\n", binName, info.CurrentVersion, info.LatestVersion)
	fmt.Fprintf(w, "Update with: %s\n", updateCommand)
	fmt.Fprintf(w, "Release notes: %s\n", info.ReleaseURL)
}

// --- Semver parsing and comparison ---

type semver struct {
	Major int
	Minor int
	Patch int
}

func parseSemver(v string) *semver {
	if !IsReleaseVersion(v) {
		return nil
	}
	v = NormalizeVersion(v)
	// Pre-release and build suffixes are ignored: a source build such as
	// v0.2.9-3-gabc1234 counts as 0.2.9.
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	var s semver
	if _, err := fmt.Sscanf(v, "%d.%d.%d", &s.Major, &s.Minor, &s.Patch); err != nil {
		return nil
	}
	return &s
}

// compareSemver returns >0 if a > b, 0 if equal, <0 if a < b.
func compareSemver(a, b *semver) int {
	if a == nil || b == nil {
		return 0
	}
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

// saveCache writes the cache through a temp file and rename, so a process
// that exits mid-write never leaves a truncated file behind.
func saveCache(configDir string, entry cacheEntry) error {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(configDir, ".update-check-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), cachePath(configDir))
}
