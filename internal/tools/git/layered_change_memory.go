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

// AppendLayeredChangeMemoryEntry stores a LayeredChangeMemory entry in a separate file
// File naming format: lcm/{unix_timestamp_ms}_{commit_hash}.yaml
func AppendLayeredChangeMemoryEntry(appPath string, entry *LayeredChangeMemoryEntry) error {
	// Create lcm directory if it doesn't exist
	lcmDir := filepath.Join(appPath, "lcm")
	if err := os.MkdirAll(lcmDir, 0755); err != nil {
		return fmt.Errorf("failed to create lcm directory: %w", err)
	}

	// Generate filename with unix timestamp in milliseconds and commit hash
	timestampMs := time.Now().UnixNano() / int64(time.Millisecond)
	filename := fmt.Sprintf("%d_%s.yaml", timestampMs, entry.Commit)
	yamlPath := filepath.Join(lcmDir, filename)

	// Marshal entry to YAML (single entry, not array)
	yamlData, err := yaml.Marshal(entry)
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

// GetAllLCMFiles returns all LCM files in the lcm directory sorted by timestamp (oldest first)
func GetAllLCMFiles(appPath string) ([]string, error) {
	lcmDir := filepath.Join(appPath, "lcm")
	
	// Check if directory exists
	if _, err := os.Stat(lcmDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	// Read all files in the directory
	entries, err := os.ReadDir(lcmDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read lcm directory: %w", err)
	}
	
	var lcmFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			lcmFiles = append(lcmFiles, filepath.Join(lcmDir, entry.Name()))
		}
	}
	
	// Sort by filename (which includes timestamp) in ascending order (oldest first)
	for i := 0; i < len(lcmFiles)-1; i++ {
		for j := i + 1; j < len(lcmFiles); j++ {
			if filepath.Base(lcmFiles[i]) > filepath.Base(lcmFiles[j]) {
				lcmFiles[i], lcmFiles[j] = lcmFiles[j], lcmFiles[i]
			}
		}
	}
	
	return lcmFiles, nil
}

// LoadLCMEntry loads a single LayeredChangeMemoryEntry from a file
func LoadLCMEntry(filePath string) (*LayeredChangeMemoryEntry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	var entry LayeredChangeMemoryEntry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML from %s: %w", filePath, err)
	}
	
	return &entry, nil
}

// LoadAllLCMEntries loads all LCM entries from the lcm directory
func LoadAllLCMEntries(appPath string) ([]LayeredChangeMemoryEntry, error) {
	files, err := GetAllLCMFiles(appPath)
	if err != nil {
		return nil, err
	}
	
	var entries []LayeredChangeMemoryEntry
	for _, file := range files {
		entry, err := LoadLCMEntry(file)
		if err != nil {
			// Log error but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: failed to load %s: %v\n", file, err)
			continue
		}
		entries = append(entries, *entry)
	}
	
	return entries, nil
}

// MigrateOldLCMFile migrates the old .layered_change_memory.yaml file to the new lcm directory structure
func MigrateOldLCMFile(appPath string) error {
	oldPath := filepath.Join(appPath, ".layered_change_memory.yaml")
	
	// Check if old file exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		// No old file to migrate
		return nil
	}
	
	// Read the old file
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("failed to read old LCM file: %w", err)
	}
	
	// Parse YAML
	var entries []LayeredChangeMemoryEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to parse old LCM file: %w", err)
	}
	
	// Create lcm directory
	lcmDir := filepath.Join(appPath, "lcm")
	if err := os.MkdirAll(lcmDir, 0755); err != nil {
		return fmt.Errorf("failed to create lcm directory: %w", err)
	}
	
	// Migrate each entry to a separate file
	for i, entry := range entries {
		// Parse timestamp to get unix milliseconds
		parsedTime, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			// If parsing fails, use index-based timestamp
			parsedTime = time.Now().Add(time.Duration(-len(entries)+i) * time.Second)
		}
		timestampMs := parsedTime.UnixNano() / int64(time.Millisecond)
		
		// Generate filename
		filename := fmt.Sprintf("%d_%s.yaml", timestampMs, entry.Commit)
		filePath := filepath.Join(lcmDir, filename)
		
		// Marshal single entry to YAML
		yamlData, err := yaml.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal entry %d: %w", i, err)
		}
		
		// Write to file
		if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
			return fmt.Errorf("failed to write entry %d to file: %w", i, err)
		}
	}
	
	// Rename old file to indicate it has been migrated
	backupPath := oldPath + ".migrated"
	if err := os.Rename(oldPath, backupPath); err != nil {
		return fmt.Errorf("failed to rename old LCM file: %w", err)
	}
	
	fmt.Fprintf(os.Stderr, "Successfully migrated %d entries from %s to lcm directory\n", len(entries), oldPath)
	fmt.Fprintf(os.Stderr, "Old file backed up to: %s\n", backupPath)
	
	return nil
}
