package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// enhancePM2Output adds port information to PM2 output
func enhancePM2Output(output string, appsDir string) string {
	// Add graceful degradation - if parsing fails, return original output
	defer func() {
		if r := recover(); r != nil {
			// If anything goes wrong, just return the original output
			fmt.Fprintf(os.Stderr, "Warning: failed to enhance PM2 output: %v\n", r)
		}
	}()
	
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
			// Add port column to header - check if cpu column exists
			if strings.Contains(line, "│ cpu │") {
				enhancedLines = append(enhancedLines, strings.Replace(line, "│ cpu │", "│ port │ cpu │", 1))
			} else {
				// If no cpu column, add port column after status
				enhancedLines = append(enhancedLines, strings.Replace(line, "│ status │", "│ status │ port │", 1))
			}
			continue
		}
		
		// Check for the separator line after header
		if inTable && i == headerIndex+1 && strings.Contains(line, "├────") {
			// Add separator for port column
			if strings.Contains(line, "├─────┼") {
				enhancedLines = append(enhancedLines, strings.Replace(line, "├─────┼", "├──────┼─────┼", 1))
			} else {
				// Just add the line as-is if format is different
				enhancedLines = append(enhancedLines, line)
			}
			continue
		}
		
		// Process data rows
		if inTable && strings.HasPrefix(line, "│") && !strings.Contains(line, "└────") {
			// Parse the app name from the line
			parts := strings.Split(line, "│")
			if len(parts) > 2 {
				appName := strings.TrimSpace(parts[2])
				if appName != "" && appName != "name" {
					// Get port for this app - with error handling
					appPath := filepath.Join(appsDir, appName)
					port := 0
					func() {
						defer func() {
							if r := recover(); r != nil {
								port = 0 // Default to 0 if port detection fails
							}
						}()
						port = GetAppPort(appPath)
					}()
					
					// Insert port column
					// Find position after status column
					statusIndex := 3
					if len(parts) > statusIndex+1 {
						newParts := make([]string, 0, len(parts)+1)
						newParts = append(newParts, parts[:statusIndex+1]...)
						if port > 0 {
							newParts = append(newParts, fmt.Sprintf(" %d ", port))
						} else {
							newParts = append(newParts, " - ")
						}
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