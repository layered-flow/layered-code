package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/tools/git"
	"gopkg.in/yaml.v3"
)

func TestLcmList(t *testing.T) {
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
	result, err := LcmList(testApp)
	if err != nil {
		t.Fatalf("Failed to list LCM entries: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success=true, got false")
	}
	if len(result.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(result.Entries))
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

	// Test 3: List entries
	result, err = LcmList(testApp)
	if err != nil {
		t.Fatalf("Failed to list LCM entries: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success=true, got false")
	}
	if len(result.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(result.Entries))
	}
	if result.Total != 3 {
		t.Errorf("Expected total=3, got %d", result.Total)
	}

	// Verify entries content
	for i, entry := range result.Entries {
		if entry.Index != i {
			t.Errorf("Expected index=%d, got %d", i, entry.Index)
		}
		if entry.Timestamp != entries[i].Timestamp {
			t.Errorf("Entry %d: Expected timestamp=%s, got %s", i, entries[i].Timestamp, entry.Timestamp)
		}
		if entry.Commit != entries[i].Commit {
			t.Errorf("Entry %d: Expected commit=%s, got %s", i, entries[i].Commit, entry.Commit)
		}
		if entry.Summary != entries[i].Summary {
			t.Errorf("Entry %d: Expected summary=%s, got %s", i, entries[i].Summary, entry.Summary)
		}
	}

	// Test 4: Invalid app name
	_, err = LcmList("")
	if err == nil {
		t.Error("Expected error for empty app name")
	}
}