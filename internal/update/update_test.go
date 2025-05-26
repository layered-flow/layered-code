package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"current older major", "1.0.0", "2.0.0", -1},
		{"current newer major", "2.0.0", "1.0.0", 1},
		{"current older minor", "1.0.0", "1.1.0", -1},
		{"current newer minor", "1.1.0", "1.0.0", 1},
		{"current older patch", "1.0.0", "1.0.1", -1},
		{"current newer patch", "1.0.1", "1.0.0", 1},
		{"different lengths", "1.0", "1.0.0", 0},
		{"current shorter", "1.0", "1.0.1", -1},
		{"latest shorter", "1.0.1", "1.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.current, tt.latest)
			if result != tt.expected {
				t.Errorf("compareVersions(%s, %s) = %d; want %d", tt.current, tt.latest, result, tt.expected)
			}
		})
	}
}

func getTestCacheDir() (string, error) {
	// Use XDG Base Directory Specification (same logic as getCacheDir)
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

func TestCheckForUpdate(t *testing.T) {
	// Clean up any existing cache for test
	cacheDir, _ := getTestCacheDir()
	cachePath := filepath.Join(cacheDir, "update_check.json")
	os.Remove(cachePath)

	// Test with dev version
	hasUpdate, _, err := CheckForUpdate("dev")
	if err != nil {
		t.Errorf("CheckForUpdate(dev) returned error: %v", err)
	}
	if hasUpdate {
		t.Error("CheckForUpdate(dev) should not indicate an update is available")
	}
}

func TestCacheReadWrite(t *testing.T) {
	// Clean up any existing cache
	cacheDir, _ := getTestCacheDir()
	cachePath := filepath.Join(cacheDir, "update_check.json")
	os.Remove(cachePath)
	defer os.Remove(cachePath)

	// Test writing cache
	cache := &UpdateCache{
		LastChecked:   time.Now(),
		LatestVersion: "v1.2.3",
		HasUpdate:     true,
	}

	err := writeCache(cache)
	if err != nil {
		t.Fatalf("Failed to write cache: %v", err)
	}

	// Test reading cache
	readCacheData, err := readCache()
	if err != nil {
		t.Fatalf("Failed to read cache: %v", err)
	}

	if readCacheData.LatestVersion != cache.LatestVersion {
		t.Errorf("Cache version mismatch: got %s, want %s", readCacheData.LatestVersion, cache.LatestVersion)
	}

	if readCacheData.HasUpdate != cache.HasUpdate {
		t.Errorf("Cache hasUpdate mismatch: got %v, want %v", readCacheData.HasUpdate, cache.HasUpdate)
	}
}
