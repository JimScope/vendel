package services

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	githubOwner     = "JimScope"
	githubRepo      = "vendel"
	githubAPIBase   = "https://api.github.com"
	latestCacheTTL  = 6 * time.Hour
	backupSuffix    = ".backup"
	downloadTimeout = 5 * time.Minute
	apiTimeout      = 30 * time.Second
	maxBinarySize   = 500 << 20 // 500 MB safety cap
)

// ReleaseInfo holds metadata about a GitHub release.
type ReleaseInfo struct {
	Version    string
	AssetURL   string // download URL for the platform-specific archive
	Checksum   string // SHA256 from checksums.txt
	CheckedAt  time.Time
	ReleaseURL string // HTML URL for release notes
}

var (
	cachedRelease *ReleaseInfo
	cacheMu       sync.RWMutex
	apiClient     = &http.Client{Timeout: apiTimeout}
)

// CheckLatest fetches the latest release from GitHub, with a 6-hour cache.
func CheckLatest(currentVersion string) (*ReleaseInfo, error) {
	cacheMu.RLock()
	if cachedRelease != nil && time.Since(cachedRelease.CheckedAt) < latestCacheTTL {
		r := *cachedRelease
		cacheMu.RUnlock()
		return &r, nil
	}
	cacheMu.RUnlock()

	release, err := fetchLatestRelease()
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	cachedRelease = release
	cacheMu.Unlock()

	return release, nil
}

// InvalidateCache forces the next CheckLatest to fetch from GitHub.
func InvalidateCache() {
	cacheMu.Lock()
	cachedRelease = nil
	cacheMu.Unlock()
}

func fetchLatestRelease() (*ReleaseInfo, error) {
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("self-update is not supported on Windows")
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPIBase, githubOwner, githubRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "vendel-updater")

	resp, err := apiClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var data struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	version := strings.TrimPrefix(data.TagName, "v")
	platform := DetectPlatform()
	archiveName := fmt.Sprintf("vendel_%s.tar.gz", platform)

	var assetURL, checksumURL string
	for _, a := range data.Assets {
		if a.Name == archiveName {
			assetURL = a.BrowserDownloadURL
		}
		if a.Name == "checksums.txt" {
			checksumURL = a.BrowserDownloadURL
		}
	}

	if assetURL == "" {
		return nil, fmt.Errorf("no asset found for platform %s in release %s", platform, version)
	}

	var checksum string
	if checksumURL != "" {
		checksum, _ = fetchChecksum(checksumURL, archiveName)
	}

	return &ReleaseInfo{
		Version:    version,
		AssetURL:   assetURL,
		Checksum:   checksum,
		CheckedAt:  time.Now(),
		ReleaseURL: data.HTMLURL,
	}, nil
}

func fetchChecksum(checksumURL, archiveName string) (string, error) {
	resp, err := apiClient.Get(checksumURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(body), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == archiveName {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", archiveName)
}

// DetectPlatform returns the GoReleaser-style platform string.
func DetectPlatform() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}

// DownloadAndVerify downloads the archive, verifies its SHA256, and extracts the binary.
// The temp directory is created next to the current binary to avoid cross-device rename issues.
// Returns the path to the extracted binary.
func DownloadAndVerify(assetURL, expectedChecksum string) (string, error) {
	// Create temp dir next to the binary (same filesystem) to allow atomic rename
	currentBinary, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot determine binary path: %w", err)
	}
	binaryDir := filepath.Dir(currentBinary)

	tmpDir, err := os.MkdirTemp(binaryDir, ".vendel-update-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	archivePath := filepath.Join(tmpDir, "vendel.tar.gz")

	// Download
	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Get(assetURL)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("download returned %d", resp.StatusCode)
	}

	f, err := os.Create(archivePath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher)
	if _, err := io.Copy(writer, resp.Body); err != nil {
		f.Close()
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("download write failed: %w", err)
	}
	f.Close()

	// Verify checksum
	if expectedChecksum != "" {
		actual := hex.EncodeToString(hasher.Sum(nil))
		if actual != expectedChecksum {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actual)
		}
	}

	// Extract binary from tar.gz
	binaryPath, err := extractBinary(archivePath, tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	return binaryPath, nil
}

func extractBinary(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the vendel binary (skip directories and other files)
		name := filepath.Base(header.Name)
		if name == "vendel" && header.Typeflag == tar.TypeReg {
			outPath := filepath.Join(destDir, "vendel")
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, io.LimitReader(tr, maxBinarySize)); err != nil {
				out.Close()
				return "", err
			}
			if err := out.Sync(); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return outPath, nil
		}
	}

	return "", fmt.Errorf("binary 'vendel' not found in archive")
}

// ApplyUpdate replaces the current binary with a backup of the old one.
func ApplyUpdate(newBinaryPath string) error {
	currentBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine current binary path: %w", err)
	}
	currentBinary, err = filepath.EvalSymlinks(currentBinary)
	if err != nil {
		return fmt.Errorf("cannot resolve binary path: %w", err)
	}

	backupPath := currentBinary + backupSuffix

	// Backup current binary
	if err := copyFile(currentBinary, backupPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Atomic replace (same filesystem — temp dir is next to the binary)
	if err := os.Rename(newBinaryPath, currentBinary); err != nil {
		// Restore from backup on failure
		_ = os.Rename(backupPath, currentBinary)
		return fmt.Errorf("binary replace failed: %w", err)
	}

	// Clean up temp dir
	os.RemoveAll(filepath.Dir(newBinaryPath))

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
