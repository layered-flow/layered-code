package tools

import (
	"fmt"
	"path/filepath"
	"strings"
)

// enhancePM2Output adds port information to PM2 output
func enhancePM2Output(output string, appsDir string) string {
	lines := strings.Split(output, "\n")
	enhancedLines := make([]string, 0, len(lines))
	
	// Track if we're in the table section
	inTable := false
	headerIndex := -1
	
	for i, line := range lines {
		// Check if this is the table header
		if strings.Contains(line, "│ id │ name") && strings.Contains(line, "│ status") {
			inTable = true
			headerIndex = i
			// Add port column to header
			enhancedLines = append(enhancedLines, strings.Replace(line, "│ cpu │", "│ port │ cpu │", 1))
			continue
		}
		
		// Check for the separator line after header
		if inTable && i == headerIndex+1 && strings.Contains(line, "├────") {
			// Add separator for port column
			enhancedLines = append(enhancedLines, strings.Replace(line, "├─────┼", "├──────┼─────┼", 1))
			continue
		}
		
		// Process data rows
		if inTable && strings.HasPrefix(line, "│") && !strings.Contains(line, "└────") {
			// Parse the app name from the line
			parts := strings.Split(line, "│")
			if len(parts) > 2 {
				appName := strings.TrimSpace(parts[2])
				if appName != "" && appName != "name" {
					// Get port for this app
					appPath := filepath.Join(appsDir, appName)
					port := GetAppPort(appPath)
					
					// Insert port column
					// Find position after status column
					statusIndex := 3
					if len(parts) > statusIndex+1 {
						newParts := make([]string, 0, len(parts)+1)
						newParts = append(newParts, parts[:statusIndex+1]...)
						newParts = append(newParts, fmt.Sprintf(" %d ", port))
						newParts = append(newParts, parts[statusIndex+1:]...)
						line = strings.Join(newParts, "│")
					}
				}
			}
		}
		
		// Check for end of table
		if strings.Contains(line, "└────") {
			inTable = false
		}
		
		enhancedLines = append(enhancedLines, line)
	}
	
	return strings.Join(enhancedLines, "\n")
}