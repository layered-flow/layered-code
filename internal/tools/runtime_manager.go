package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"
)

// RuntimeManager handles runtime information for apps
type RuntimeManager struct {
	mu sync.RWMutex
}

var runtimeMgr = &RuntimeManager{}

// RuntimeInfo stores runtime information for an app
type RuntimeInfo struct {
	Port   int    `json:"port,omitempty"`
	Status string `json:"status,omitempty"`
}

// GetRuntimeInfo retrieves runtime information for an app
func GetRuntimeInfo(appName string) (RuntimeInfo, error) {
	runtimeMgr.mu.RLock()
	defer runtimeMgr.mu.RUnlock()

	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return RuntimeInfo{}, err
	}

	appPath := filepath.Join(appsDir, appName)
	runtimePath := filepath.Join(appPath, ".layered-code", "runtime.json")

	data, err := os.ReadFile(runtimePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default info if file doesn't exist
			return RuntimeInfo{
				Port:   0, // Don't call GetAppPort here to avoid recursion
				Status: "unknown",
			}, nil
		}
		return RuntimeInfo{}, err
	}

	var info RuntimeInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return RuntimeInfo{}, err
	}

	return info, nil
}

// SaveRuntimeInfo saves runtime information for an app
func SaveRuntimeInfo(appName string, info RuntimeInfo) error {
	runtimeMgr.mu.Lock()
	defer runtimeMgr.mu.Unlock()

	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return err
	}

	appPath := filepath.Join(appsDir, appName)
	layeredDir := filepath.Join(appPath, ".layered-code")

	if err := os.MkdirAll(layeredDir, constants.AppsDirectoryPerms); err != nil {
		return err
	}

	runtimePath := filepath.Join(layeredDir, "runtime.json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(runtimePath, data, 0644)
}

// GetAppPort attempts to determine the port for an app
func GetAppPort(appPath string) int {
	appName := filepath.Base(appPath)
	
	// First, check if there's a runtime info file (direct file read to avoid recursion)
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
	if port, err := detectPortFromPackageJSON(appPath); err == nil && port > 0 {
		return port
	}

	// Generate a unique port based on app name
	return GenerateUniquePort(appName)
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

// detectPortFromPackageJSON checks package.json for port configurations
func detectPortFromPackageJSON(appPath string) (int, error) {
	packagePath := filepath.Join(appPath, "package.json")
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return 0, err
	}

	// Look for --port flag in scripts
	portRegex := regexp.MustCompile(`--port[= ](\d+)`)
	if matches := portRegex.FindSubmatch(data); len(matches) > 1 {
		var port int
		fmt.Sscanf(string(matches[1]), "%d", &port)
		return port, nil
	}

	return 0, fmt.Errorf("no port found in package.json")
}

// GenerateUniquePort generates a unique port for an app based on its name
func GenerateUniquePort(appName string) int {
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

// CheckPortConflicts checks if the given port is already in use by another app
func CheckPortConflicts(excludeApp string, port int) (string, error) {
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == excludeApp {
			continue
		}

		appPath := filepath.Join(appsDir, entry.Name())
		appPort := GetAppPort(appPath)
		
		if appPort == port {
			return entry.Name(), nil
		}
	}

	return "", nil
}