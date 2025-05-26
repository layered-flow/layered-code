package constants

const (
	// Project information
	ProjectName = "layered-code"

	// Apps directory configuration
	DefaultAppsDirectory = "LayeredApps"
	AppsDirectoryEnvVar  = "LAYERED_APPS_DIRECTORY"

	// File permission constants
	AppsDirectoryPerms   = 0755
	OwnerWritePermission = 0200

	// File size constants
	MaxFileSize        = 1024 * 1024
	MaxFileSizeInWords = "1MB"

	// Directory traversal constants
	MaxDirectoryDepth = 10000
)

var (
	// ProjectVersion will be set at build time via ldflags
	ProjectVersion = "dev"
)
