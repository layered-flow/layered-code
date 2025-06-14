package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackageManager(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupFunc      func()
		preferred      string
		expectedPM     PackageManager
		skipIfNoPnpm   bool
	}{
		{
			name: "prefer pnpm when available",
			setupFunc: func() {
				// No lockfiles
			},
			preferred:    "",
			expectedPM:   PNPM,
			skipIfNoPnpm: true,
		},
		{
			name: "respect existing pnpm-lock.yaml",
			setupFunc: func() {
				os.WriteFile(filepath.Join(tempDir, "pnpm-lock.yaml"), []byte(""), 0644)
			},
			preferred:    "",
			expectedPM:   PNPM,
			skipIfNoPnpm: true,
		},
		{
			name: "respect existing package-lock.json",
			setupFunc: func() {
				os.WriteFile(filepath.Join(tempDir, "package-lock.json"), []byte(""), 0644)
			},
			preferred:  "",
			expectedPM: NPM,
		},
		{
			name: "respect user preference",
			setupFunc: func() {
				os.WriteFile(filepath.Join(tempDir, "pnpm-lock.yaml"), []byte(""), 0644)
			},
			preferred:  "npm",
			expectedPM: NPM,
		},
		{
			name: "fallback to default when invalid pm specified",
			setupFunc: func() {
				// When invalid PM is specified, it falls back to default priority
			},
			preferred: "invalid-pm",
			expectedPM: func() PackageManager {
				// This matches the actual behavior: pnpm is preferred if available
				if isPackageManagerAvailable("pnpm") {
					return PNPM
				}
				return NPM
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that require pnpm if it's not available
			if tt.skipIfNoPnpm && !isPackageManagerAvailable("pnpm") {
				t.Skip("pnpm not available")
			}

			// Clean up temp dir
			files, _ := os.ReadDir(tempDir)
			for _, f := range files {
				os.Remove(filepath.Join(tempDir, f.Name()))
			}

			// Run setup
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Test
			result := detectPackageManager(tempDir, tt.preferred)
			if result != tt.expectedPM {
				t.Errorf("detectPackageManager() = %v, want %v", result, tt.expectedPM)
			}
		})
	}
}

func TestGetInstallCommand(t *testing.T) {
	tests := []struct {
		name       string
		pm         PackageManager
		production bool
		expected   []string
	}{
		{
			name:       "npm install",
			pm:         NPM,
			production: false,
			expected:   []string{"npm", "install"},
		},
		{
			name:       "npm install production",
			pm:         NPM,
			production: true,
			expected:   []string{"npm", "install", "--production"},
		},
		{
			name:       "pnpm install",
			pm:         PNPM,
			production: false,
			expected:   []string{"pnpm", "install"},
		},
		{
			name:       "pnpm install production",
			pm:         PNPM,
			production: true,
			expected:   []string{"pnpm", "install", "--prod"},
		},
		{
			name:       "yarn install",
			pm:         YARN,
			production: false,
			expected:   []string{"yarn", "install"},
		},
		{
			name:       "yarn install production",
			pm:         YARN,
			production: true,
			expected:   []string{"yarn", "install", "--production"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInstallCommand(tt.pm, tt.production)
			if len(result) != len(tt.expected) {
				t.Errorf("getInstallCommand() returned %d args, want %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("getInstallCommand()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestIsPackageManagerAvailable(t *testing.T) {
	// npm should always be available in CI environments
	if !isPackageManagerAvailable("npm") {
		t.Error("npm should be available")
	}

	// Test with an invalid package manager
	if isPackageManagerAvailable("invalid-package-manager-12345") {
		t.Error("invalid package manager should not be available")
	}
}