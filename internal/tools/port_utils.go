package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RuntimeInfo stores runtime information for an app
type RuntimeInfo struct {
	Port   int    `json:"port,omitempty"`
	Status string `json:"status,omitempty"`
}

// detectVitePort reads vite.config.js and extracts the configured port
func detectVitePort(appPath string) (int, error) {
	// Check for vite.config.js
	viteConfigPath := filepath.Join(appPath, "vite.config.js")
	content, err := os.ReadFile(viteConfigPath)
	if err != nil {
		// Try vite.config.ts
		viteConfigPath = filepath.Join(appPath, "vite.config.ts")
		content, err = os.ReadFile(viteConfigPath)
		if err != nil {
			return 0, fmt.Errorf("vite config not found")
		}
	}

	// Look for port configuration
	// Match patterns like: port: 3000, port:3000, "port": 3000
	portRegex := regexp.MustCompile(`(?:"|')?port(?:"|')?\s*:\s*(\d+)`)
	matches := portRegex.FindSubmatch(content)
	if len(matches) > 1 {
		var port int
		fmt.Sscanf(string(matches[1]), "%d", &port)
		if port > 0 {
			return port, nil
		}
	}

	// Default Vite port
	return 5173, nil
}

// getAppPort attempts to determine the port for an app
func getAppPort(appPath string) int {
	// First, check if there's a runtime info file
	runtimePath := filepath.Join(appPath, ".layered-code", "runtime.json")
	if data, err := os.ReadFile(runtimePath); err == nil {
		var info RuntimeInfo
		if json.Unmarshal(data, &info) == nil && info.Port > 0 {
			return info.Port
		}
	}

	// Try to detect from vite config
	if port, err := detectVitePort(appPath); err == nil {
		return port
	}

	// Check package.json for scripts that might indicate port
	packagePath := filepath.Join(appPath, "package.json")
	if data, err := os.ReadFile(packagePath); err == nil {
		// Look for --port flag in scripts
		portRegex := regexp.MustCompile(`--port[= ](\d+)`)
		if matches := portRegex.FindSubmatch(data); len(matches) > 1 {
			var port int
			fmt.Sscanf(string(matches[1]), "%d", &port)
			if port > 0 {
				return port
			}
		}
	}

	// Default to 3000 (common default)
	return 3000
}

// generateUniquePort generates a unique port for an app based on its name
func generateUniquePort(appName string) int {
	// Base port
	basePort := 3000
	
	// Generate a hash from the app name
	hash := 0
	for _, char := range appName {
		hash = (hash * 31 + int(char)) % 1000
	}
	
	// Return a port in the range 3000-3999
	return basePort + hash
}

// saveRuntimeInfo saves runtime information for an app
func saveRuntimeInfo(appPath string, info RuntimeInfo) error {
	layeredDir := filepath.Join(appPath, ".layered-code")
	if err := os.MkdirAll(layeredDir, 0755); err != nil {
		return err
	}

	runtimePath := filepath.Join(layeredDir, "runtime.json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(runtimePath, data, 0644)
}

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
					port := getAppPort(appPath)
					
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