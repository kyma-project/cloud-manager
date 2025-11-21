package gcpnfsvolumerestore

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

func getLeaseName(resourceName, prefix string) string {
	if prefix != "" {
		return fmt.Sprintf("%s-%s", prefix, resourceName)
	}
	return resourceName
}
func getHolderName(ownerName types.NamespacedName) string {
	return fmt.Sprintf("%s/%s", ownerName.Namespace, ownerName.Name)
}

// convertBackupUrlToFullPath converts backup URL from format "{location_id}/{backup_id}"
// to GCP full path format "projects/{project}/locations/{location_id}/backups/{backup_id}"
func convertBackupUrlToFullPath(project, backupUrl string) (string, error) {
	parts := strings.Split(backupUrl, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid backup URL format, expected '{location_id}/{backup_id}', got '%s'", backupUrl)
	}

	locationId := parts[0]
	backupId := parts[1]

	if locationId == "" || backupId == "" {
		return "", fmt.Errorf("location_id and backup_id cannot be empty in backup URL '%s'", backupUrl)
	}

	return fmt.Sprintf("projects/%s/locations/%s/backups/%s", project, locationId, backupId), nil
}

// extractBackupLocation extracts just the location_id from the full backup name
// in format: projects/{project_number}/locations/{location_id}/backups/{backup_id}
// Returns the location_id or empty string if format is invalid
func extractBackupLocation(backupUrl string) string {
	parts := strings.Split(backupUrl, "/")
	if len(parts) != 2 {
		return ""
	}

	return parts[0]
}

// extractBackupLocation extracts just the location_id from the full backup name
// in format: projects/{project_number}/locations/{location_id}/backups/{backup_id}
// Returns the location_id or empty string if format is invalid
func extractBackupName(backupUrl string) string {
	parts := strings.Split(backupUrl, "/")
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}
