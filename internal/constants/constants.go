package constants

const (
	// Project information
	ProjectName    = "layered-code"
	ProjectVersion = "0.0.1"

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
