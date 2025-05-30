package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/tools/git"
	"gopkg.in/yaml.v3"
)

func TestLcmSearch(t *testing.T) {
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
	result, err := LcmSearch(testApp, "test", false, 10, "")
	if err != nil {
		t.Fatalf("Failed to search LCM entries: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success=true, got false")
	}
	if len(result.Matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(result.Matches))
	}
	if result.Message != "No layered change memory found" {
		t.Errorf("Expected 'No layered change memory found' message, got: %s", result.Message)
	}

	// Test 2: Create some LCM entries for searching
	lcmPath := filepath.Join(appPath, ".layered_change_memory.yaml")
	entries := []git.LayeredChangeMemoryEntry{
		{
			Timestamp:      "2024-01-01T10:00:00Z",
			Commit:         "abc1234",
			CommitMessage:  "Add contact page",
			Summary:        "Implemented contact page with form validation",
			Considerations: []string{"User rejected complex CAPTCHA", "Added simple math question instead"},
			FollowUp:       "Monitor spam submissions",
		},
		{
			Timestamp:     "2024-01-02T11:00:00Z",
			Commit:        "def5678",
			CommitMessage: "Improve performance",
			Summary:       "Optimized database queries for better performance",
			FollowUp:      "Profile under high load",
		},
		{
			Timestamp:      "2024-01-03T12:00:00Z",
			Commit:         "ghi9012",
			CommitMessage:  "Fix contact form bug",
			Summary:        "Fixed validation bug in contact form",
			Considerations: []string{"Considered removing feature but users need it"},
		},
		{
			Timestamp:     "2024-01-04T13:00:00Z",
			Commit:        "jkl3456",
			CommitMessage: "Update user profile",
			Summary:       "Added avatar upload to user profile",
			FollowUp:      "Add image size validation",
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

	// Test 3: Search for "contact" (case insensitive)
	result, err = LcmSearch(testApp, "contact", false, 10, "")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success=true")
	}
	if len(result.Matches) != 2 {
		t.Errorf("Expected 2 matches for 'contact', got %d", len(result.Matches))
	}
	if result.Matches[0].Index != 0 || result.Matches[1].Index != 2 {
		t.Errorf("Expected matches at indices 0 and 2")
	}

	// Test 4: Search for "performance" 
	result, err = LcmSearch(testApp, "performance", false, 10, "")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	if len(result.Matches) != 1 {
		t.Errorf("Expected 1 match for 'performance', got %d", len(result.Matches))
	}
	if result.Matches[0].Index != 1 {
		t.Errorf("Expected match at index 1")
	}

	// Test 5: Search for "rejected" in considerations only
	result, err = LcmSearch(testApp, "rejected", false, 10, "considerations")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	if len(result.Matches) != 1 {
		t.Errorf("Expected 1 match for 'rejected' in considerations, got %d", len(result.Matches))
	}
	if result.Matches[0].Index != 0 {
		t.Errorf("Expected match at index 0")
	}

	// Test 6: Case sensitive search
	result, err = LcmSearch(testApp, "Contact", true, 10, "")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	if len(result.Matches) != 0 {
		t.Errorf("Expected 0 matches for case-sensitive 'Contact', got %d", len(result.Matches))
	}

	// Test 7: Search in follow_up field only
	result, err = LcmSearch(testApp, "validation", false, 10, "follow_up")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	if len(result.Matches) != 1 {
		t.Errorf("Expected 1 match in follow_up, got %d", len(result.Matches))
	}
	if result.Matches[0].Index != 3 {
		t.Errorf("Expected match at index 3")
	}

	// Test 8: Max results limit
	result, err = LcmSearch(testApp, "a", false, 2, "")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	if len(result.Matches) != 2 {
		t.Errorf("Expected 2 matches (max limit), got %d", len(result.Matches))
	}
	if !result.MaxExceeded {
		t.Errorf("Expected MaxExceeded=true")
	}

	// Test 9: Empty pattern
	_, err = LcmSearch(testApp, "", false, 10, "")
	if err == nil {
		t.Error("Expected error for empty pattern")
	}

	// Test 10: Invalid field filter
	_, err = LcmSearch(testApp, "test", false, 10, "invalid_field")
	if err == nil {
		t.Error("Expected error for invalid field filter")
	}

	// Test 11: Invalid app name
	_, err = LcmSearch("", "test", false, 10, "")
	if err == nil {
		t.Error("Expected error for empty app name")
	}

	// Test 12: Search in commit_message field
	result, err = LcmSearch(testApp, "Fix", false, 10, "commit_message")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	if len(result.Matches) != 1 {
		t.Errorf("Expected 1 match in commit_message, got %d", len(result.Matches))
	}
	if result.Matches[0].Index != 2 {
		t.Errorf("Expected match at index 2")
	}

	// Test 13: Verify match context
	result, err = LcmSearch(testApp, "validation", false, 10, "")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	if len(result.Matches) < 1 {
		t.Fatal("Expected at least 1 match")
	}
	if result.Matches[0].Context == "" {
		t.Error("Expected context to be populated")
	}

	// Test 14: Verify matched fields
	result, err = LcmSearch(testApp, "contact", false, 10, "")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
	if len(result.Matches) < 1 {
		t.Fatal("Expected at least 1 match")
	}
	foundSummaryMatch := false
	foundCommitMessageMatch := false
	for _, match := range result.Matches {
		for _, field := range match.MatchedFields {
			if field == "summary" {
				foundSummaryMatch = true
			}
			if field == "commit_message" {
				foundCommitMessageMatch = true
			}
		}
	}
	if !foundSummaryMatch {
		t.Error("Expected to find match in summary field")
	}
	if !foundCommitMessageMatch {
		t.Error("Expected to find match in commit_message field")
	}
}