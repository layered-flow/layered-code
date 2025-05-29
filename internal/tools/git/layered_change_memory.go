package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LayeredChangeMemoryEntry represents a single commit entry in the LayeredChangeMemory YAML log
type LayeredChangeMemoryEntry struct {
	Timestamp      string   `yaml:"timestamp"`
	Commit         string   `yaml:"commit"`
	CommitMessage  string   `yaml:"commit_message"`
	Summary        string   `yaml:"summary"`
	Considerations []string `yaml:"considerations,omitempty"`
	FollowUp       string   `yaml:"follow_up,omitempty"`
}

// GenerateLayeredChangeMemoryEntry creates a LayeredChangeMemory entry for a git commit
func GenerateLayeredChangeMemoryEntry(appPath string, commitHash string, commitMessage string, summary string, considerations []string, followUp string) (*LayeredChangeMemoryEntry, error) {
	// Get the current timestamp
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Truncate commit hash to short version if needed
	if len(commitHash) > 7 {
		commitHash = commitHash[:7]
	}

	// Limit commit message to max 2 lines
	lines := strings.Split(commitMessage, "\n")
	if len(lines) > 2 {
		commitMessage = strings.Join(lines[:2], "\n")
	}

	// Ensure considerations are limited to 3 items
	if len(considerations) > 3 {
		considerations = considerations[:3]
	}

	entry := &LayeredChangeMemoryEntry{
		Timestamp:      timestamp,
		Commit:         commitHash,
		CommitMessage:  commitMessage,
		Summary:        summary,
		Considerations: considerations,
		FollowUp:       followUp,
	}

	return entry, nil
}

// AppendLayeredChangeMemoryEntry appends a LayeredChangeMemory entry to the .layered_change_memory.yaml file
func AppendLayeredChangeMemoryEntry(appPath string, entry *LayeredChangeMemoryEntry) error {
	yamlPath := filepath.Join(appPath, ".layered_change_memory.yaml")

	// Read existing entries
	var entries []LayeredChangeMemoryEntry
	if data, err := os.ReadFile(yamlPath); err == nil {
		if err := yaml.Unmarshal(data, &entries); err != nil {
			// If unmarshaling fails, start fresh
			entries = []LayeredChangeMemoryEntry{}
		}
	}

	// Append new entry
	entries = append(entries, *entry)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(yamlPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}

// GetCommitDetails retrieves detailed information about a commit
func GetCommitDetails(appPath string, commitHash string) (map[string]interface{}, error) {
	details := make(map[string]interface{})

	// Get commit message
	msgCmd := exec.Command("git", "show", "-s", "--format=%B", commitHash)
	msgCmd.Dir = appPath
	msgOutput, err := msgCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit message: %w", err)
	}
	details["message"] = strings.TrimSpace(string(msgOutput))

	// Get author name
	authorCmd := exec.Command("git", "show", "-s", "--format=%an", commitHash)
	authorCmd.Dir = appPath
	authorOutput, err := authorCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	details["author"] = strings.TrimSpace(string(authorOutput))

	// Get files changed
	filesCmd := exec.Command("git", "show", "--name-status", "--format=", commitHash)
	filesCmd.Dir = appPath
	filesOutput, err := filesCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	// Parse file changes
	var modifiedFiles []string
	var addedFiles []string
	var deletedFiles []string

	lines := strings.Split(string(filesOutput), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			status := parts[0]
			filename := parts[1]

			switch status {
			case "A":
				addedFiles = append(addedFiles, filename)
			case "M":
				modifiedFiles = append(modifiedFiles, filename)
			case "D":
				deletedFiles = append(deletedFiles, filename)
			}
		}
	}

	details["added_files"] = addedFiles
	details["modified_files"] = modifiedFiles
	details["deleted_files"] = deletedFiles

	// Get diff statistics
	statCmd := exec.Command("git", "show", "--stat", "--format=", commitHash)
	statCmd.Dir = appPath
	statOutput, err := statCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff stats: %w", err)
	}
	details["stats"] = strings.TrimSpace(string(statOutput))

	return details, nil
}
