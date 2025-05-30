package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/tools/git"
	"gopkg.in/yaml.v3"
)

func TestLcmRead(t *testing.T) {
	// Set up test environment
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	testApp := "testapp"
	appPath := filepath.Join(appsDir, testApp)
	os.MkdirAll(appPath, 0755)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	// Test 1: No LCM file exists
	result, err := LcmRead(testApp, 0)
	if err != nil {
		t.Fatalf("Failed to read LCM entry: %v", err)
	}
	if result.Success {
		t.Errorf("Expected success=false when no LCM file exists")
	}
	if result.Message != "No layered change memory found" {
		t.Errorf("Expected 'No layered change memory found' message, got: %s", result.Message)
	}

	// Test 2: Create some LCM entries
	lcmPath := filepath.Join(appPath, ".layered_change_memory.yaml")

	// Create test entries
	entries := []git.LayeredChangeMemoryEntry{
		{
			Timestamp:      "2024-01-01T10:00:00Z",
			Commit:         "abc1234",
			CommitMessage:  "Initial commit",
			Summary:        "Set up project structure",
			Considerations: []string{"Used standard layout", "Added basic tests"},
			FollowUp:       "Add more comprehensive tests",
		},
		{
			Timestamp:     "2024-01-02T11:00:00Z",
			Commit:        "def5678",
			CommitMessage: "Add feature X",
			Summary:       "Implemented feature X with full test coverage",
			FollowUp:      "Consider performance optimizations",
		},
		{
			Timestamp:      "2024-01-03T12:00:00Z",
			Commit:         "ghi9012",
			CommitMessage:  "Fix bug in feature Y",
			Summary:        "Fixed critical bug that caused data loss",
			Considerations: []string{"Added validation", "Updated error handling"},
		},
	}

	// Write entries to YAML file
	yamlData, err := yaml.Marshal(entries)
	if err != nil {
		t.Fatalf("Failed to marshal test entries: %v", err)
	}
	if err := os.WriteFile(lcmPath, yamlData, 0644); err != nil {
		t.Fatalf("Failed to write test LCM file: %v", err)
	}

	// Test 3: Read valid entry (index 0)
	result, err = LcmRead(testApp, 0)
	if err != nil {
		t.Fatalf("Failed to read LCM entry: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success=true, got false")
	}
	if result.Entry == nil {
		t.Fatal("Expected entry to be non-nil")
	}
	if result.Entry.Timestamp != entries[0].Timestamp {
		t.Errorf("Expected timestamp=%s, got %s", entries[0].Timestamp, result.Entry.Timestamp)
	}
	if result.Entry.Commit != entries[0].Commit {
		t.Errorf("Expected commit=%s, got %s", entries[0].Commit, result.Entry.Commit)
	}
	if result.Entry.Summary != entries[0].Summary {
		t.Errorf("Expected summary=%s, got %s", entries[0].Summary, result.Entry.Summary)
	}
	if len(result.Entry.Considerations) != 2 {
		t.Errorf("Expected 2 considerations, got %d", len(result.Entry.Considerations))
	}

	// Test 4: Read valid entry (index 2)
	result, err = LcmRead(testApp, 2)
	if err != nil {
		t.Fatalf("Failed to read LCM entry: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success=true, got false")
	}
	if result.Entry == nil {
		t.Fatal("Expected entry to be non-nil")
	}
	if result.Entry.Commit != entries[2].Commit {
		t.Errorf("Expected commit=%s, got %s", entries[2].Commit, result.Entry.Commit)
	}

	// Test 5: Invalid index (negative)
	result, err = LcmRead(testApp, -1)
	if err != nil {
		t.Fatalf("Failed to read LCM entry: %v", err)
	}
	if result.Success {
		t.Errorf("Expected success=false for negative index")
	}
	if result.Message != "Invalid index -1. Valid range is 0-2" {
		t.Errorf("Expected invalid index message, got: %s", result.Message)
	}

	// Test 6: Invalid index (too large)
	result, err = LcmRead(testApp, 10)
	if err != nil {
		t.Fatalf("Failed to read LCM entry: %v", err)
	}
	if result.Success {
		t.Errorf("Expected success=false for out-of-range index")
	}
	if result.Message != "Invalid index 10. Valid range is 0-2" {
		t.Errorf("Expected invalid index message, got: %s", result.Message)
	}

	// Test 7: Invalid app name
	_, err = LcmRead("", 0)
	if err == nil {
		t.Error("Expected error for empty app name")
	}
}