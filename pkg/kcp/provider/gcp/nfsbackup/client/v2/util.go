package v2

import (
	"fmt"
	"regexp"
)

// fileBackupPattern matches: projects/{project}/locations/{location}/backups/{backupName}
var fileBackupPathRegex = regexp.MustCompile(`^projects/([^/]+)/locations/([^/]+)/backups/([^/]+)$`)

// GetFileBackupPath returns the full GCP Filestore backup resource name.
// Format: projects/{project}/locations/{location}/backups/{name}
func GetFileBackupPath(projectId, location, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
}

// GetFilestoreParentPath returns the parent path for Filestore resources.
// Format: projects/{project}/locations/{location}
// Use "-" for location to list across all locations.
func GetFilestoreParentPath(projectId, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", projectId, location)
}

// GetProjectLocationNameFromFileBackupPath extracts components from a backup path.
// Returns empty strings if the path doesn't match the expected format.
func GetProjectLocationNameFromFileBackupPath(fullPath string) (projectId, location, name string) {
	matches := fileBackupPathRegex.FindStringSubmatch(fullPath)
	if len(matches) != 4 {
		return "", "", ""
	}
	return matches[1], matches[2], matches[3]
}
