package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/layered-flow/layered-code/internal/constants"
)

const (
	cacheFileName = "update_check.json"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type UpdateCache struct {
	LastChecked   time.Time `json:"last_checked"`
	LatestVersion string    `json:"latest_version"`
	HasUpdate     bool      `json:"has_update"`
}

func getCacheDir() (string, error) {
	// Use XDG Base Directory Specification
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheHome = filepath.Join(homeDir, ".cache")
	}
	return filepath.Join(cacheHome, "layered-code"), nil
}

func getCachePath() (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, cacheFileName), nil
}

func readCache() (*UpdateCache, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache UpdateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func writeCache(cache *UpdateCache) error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

func CheckForUpdate(currentVersion string) (bool, string, error) {
	if currentVersion == "dev" {
		return false, "", nil
	}

	// Check if we have a recent cache
	cache, err := readCache()
	if err == nil && time.Since(cache.LastChecked) < constants.UpdateCheckInterval {
		return cache.HasUpdate, cache.LatestVersion, nil
	}

	// Perform fresh check
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(constants.GitHubApiRepoUrl + "/releases/latest")
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("failed to fetch latest release: status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, "", err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	hasUpdate := compareVersions(currentVersion, latestVersion) < 0

	// Update cache
	newCache := &UpdateCache{
		LastChecked:   time.Now(),
		LatestVersion: release.TagName,
		HasUpdate:     hasUpdate,
	}

	// Ignore cache write errors - they shouldn't prevent the check from succeeding
	_ = writeCache(newCache)

	return hasUpdate, release.TagName, nil
}

func compareVersions(current, latest string) int {
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Pad shorter version with zeros
	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for i := 0; i < maxLen; i++ {
		var curr, lat int

		if i < len(currentParts) {
			fmt.Sscanf(currentParts[i], "%d", &curr)
		}

		if i < len(latestParts) {
			fmt.Sscanf(latestParts[i], "%d", &lat)
		}

		if curr < lat {
			return -1
		} else if curr > lat {
			return 1
		}
	}

	return 0
}

func DisplayUpdateWarning(latestVersion string) {
	fmt.Printf("\n⚠️  Update available: %s\n", latestVersion)
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		fmt.Println("Update via Homebrew: brew update && brew upgrade layered-code")
		fmt.Printf("Or download from: %s/releases/latest\n", constants.GitHubRepoUrl)
	} else {
		fmt.Printf("Download the latest version from: %s/releases/latest\n", constants.GitHubRepoUrl)
	}
	fmt.Println()
}
