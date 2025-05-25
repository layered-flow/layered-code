package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/constants"
)

func TestGetAppsDirectory(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		envVar  string
		want    string
		wantErr bool
	}{
		{"", filepath.Join(homeDir, constants.DefaultAppsDirectory), false},
		{"MyApps", filepath.Join(homeDir, "MyApps"), false},
		{"~/CustomApps", filepath.Join(homeDir, "CustomApps"), false},
		{"../../../etc", "", true},
		{"/tmp/apps", "", true},
	}

	for _, tt := range tests {
		originalEnv := os.Getenv(constants.AppsDirectoryEnvVar)
		os.Setenv(constants.AppsDirectoryEnvVar, tt.envVar)

		got, err := GetAppsDirectory()

		if (err != nil) != tt.wantErr {
			t.Errorf("envVar=%q: wantErr=%v, got=%v", tt.envVar, tt.wantErr, err)
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("envVar=%q: got=%q, want=%q", tt.envVar, got, tt.want)
		}

		os.Setenv(constants.AppsDirectoryEnvVar, originalEnv)
	}
}

func TestValidateAppsDirectoryPath(t *testing.T) {
	homeDir := "/Users/testuser"

	tests := []struct {
		path    string
		wantErr error
	}{
		{"MyApps", nil},
		{"../../../etc", ErrDirectoryTraversal},
		{"/tmp/apps", ErrAbsolutePathNotInHome},
		{filepath.Join(homeDir, "apps"), nil},
	}

	for _, tt := range tests {
		err := validateAppsDirectoryPath(tt.path, homeDir)
		if err != tt.wantErr {
			t.Errorf("path=%q: got=%v, want=%v", tt.path, err, tt.wantErr)
		}
	}
}

func TestIsWithinDirectory(t *testing.T) {
	tests := []struct {
		target, base string
		want         bool
	}{
		{"/home/user/apps", "/home/user", true},
		{"/tmp/apps", "/home/user", false},
		{"/home/user", "/home/user", true},
		{"/home/user/../etc", "/home/user", false},
	}

	for _, tt := range tests {
		got := IsWithinDirectory(tt.target, tt.base)
		if got != tt.want {
			t.Errorf("IsWithinDirectory(%q, %q) = %v, want %v", tt.target, tt.base, got, tt.want)
		}
	}
}

func TestCheckDirectoryWritable(t *testing.T) {
	tmpDir := t.TempDir()

	// Test writable directory
	if err := checkDirectoryWritable(tmpDir); err != nil {
		t.Errorf("writable directory failed: %v", err)
	}

	// Test file (not directory)
	tmpFile := filepath.Join(tmpDir, "testfile")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	if err := checkDirectoryWritable(tmpFile); err != ErrNotADirectory {
		t.Errorf("expected ErrNotADirectory, got: %v", err)
	}
}
