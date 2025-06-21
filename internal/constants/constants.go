package constants

import "time"

const (
	// Project information
	ProjectName      = "layered-code"
	GitHubRepoUrl    = "https://github.com/layered-flow/layered-code"
	GitHubApiRepoUrl = "https://api.github.com/repos/layered-flow/layered-code"

	// Update configuration
	UpdateCheckInterval = 24 * time.Hour

	// Apps directory configuration
	DefaultAppsDirectory = "LayeredApps"
	AppsDirectoryEnvVar  = "LAYERED_APPS_DIRECTORY"

	// File permission constants
	AppsDirectoryPerms   = 0755
	OwnerWritePermission = 0200

	// File size constants
	MaxFileSize        = 10 * 1024 * 1024 // 10MB
	MaxFileSizeInWords = "10MB"

	// Directory traversal constants
	MaxDirectoryDepth = 10000
)

var (
	// ProjectVersion will be set at build time via ldflags
	ProjectVersion = "dev"
)
