package pnpm

import (
	"strings"
	"testing"
)

func TestPnpmInstallValidation(t *testing.T) {
	// Test cases for validation only
	tests := []struct {
		name      string
		appName   string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty app name",
			appName:   "",
			wantErr:   true,
			errMsg:    "app name is required",
		},
		{
			name:      "non-existent app",
			appName:   "definitely-does-not-exist-app-12345",
			wantErr:   true,
			errMsg:    "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the function - it will fail early on validation
			_, err := PnpmInstall(tt.appName, false)

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("PnpmInstall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("PnpmInstall() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}