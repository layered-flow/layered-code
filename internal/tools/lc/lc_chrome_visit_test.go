package lc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/constants"
)

func TestLcChromeVisit(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	testAppsDir := filepath.Join(homeDir, "TestLayeredApps_ChromeVisit_"+t.Name())
	defer os.RemoveAll(testAppsDir)

	t.Setenv(constants.AppsDirectoryEnvVar, testAppsDir)

	if err := os.MkdirAll(testAppsDir, constants.AppsDirectoryPerms); err != nil {
		t.Fatalf("Failed to create test apps directory: %v", err)
	}

	appName := "test-chrome-app"
	url := "https://example.com"

	result, err := LcChromeVisit(appName, url)
	if err != nil {
		t.Logf("Chrome visit returned error (expected if Chrome not available): %v", err)
		return
	}

	if !result.Success {
		t.Logf("Chrome visit was not successful: %s", result.Message)
		return
	}

	if result.PageContent == "" {
		t.Errorf("Expected page content, got empty string")
	}

	if result.AppName != appName {
		t.Errorf("Expected app name %q, got %q", appName, result.AppName)
	}

	if result.URL != url {
		t.Errorf("Expected URL %q, got %q", url, result.URL)
	}
}

func TestLcChromeVisitInvalidAppName(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	testAppsDir := filepath.Join(homeDir, "TestLayeredApps_ChromeVisit_"+t.Name())
	defer os.RemoveAll(testAppsDir)

	t.Setenv(constants.AppsDirectoryEnvVar, testAppsDir)

	testCases := []struct {
		name    string
		appName string
		wantErr string
	}{
		{
			name:    "empty app name",
			appName: "",
			wantErr: "invalid app name: app name cannot be empty",
		},
		{
			name:    "app name with path traversal",
			appName: "../evil",
			wantErr: "invalid app name: app name cannot contain '..'",
		},
		{
			name:    "app name with slash",
			appName: "test/app",
			wantErr: "invalid app name: app name cannot contain '/'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LcChromeVisit(tc.appName, "https://example.com")
			if err == nil {
				t.Errorf("Expected error for app name %q, got nil", tc.appName)
			} else if err.Error() != tc.wantErr {
				t.Errorf("Expected error %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestLcChromeVisitEmptyURL(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	testAppsDir := filepath.Join(homeDir, "TestLayeredApps_ChromeVisit_"+t.Name())
	defer os.RemoveAll(testAppsDir)

	t.Setenv(constants.AppsDirectoryEnvVar, testAppsDir)

	_, err = LcChromeVisit("test-app", "")
	if err == nil {
		t.Error("Expected error for empty URL, got nil")
	} else if err.Error() != "URL cannot be empty" {
		t.Errorf("Expected error 'URL cannot be empty', got %q", err.Error())
	}
}