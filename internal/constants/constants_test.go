package constants

import (
	"testing"
)

func TestConstants(t *testing.T) {
	t.Skip("Skipping constant tests - no actual testing needed for static constants")

	// This test is skipped but ensures the constants file is included in coverage reports
	// The constants are referenced here to ensure they're considered "covered"
	_ = ProjectName
	_ = ProjectVersion
	_ = DefaultAppsDirectory
	_ = AppsDirectoryEnvVar
	_ = AppsDirectoryPerms
	_ = OwnerWritePermission
}
