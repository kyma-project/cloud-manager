package gcpnfsvolumebackupdiscovery

import "strings"

// extractBackupUri extracts location_id/backup_id from the full backup name
// in format: projects/{project_number}/locations/{location_id}/backups/{backup_id}
// Returns the formatted string "{location_id}/{backup_id}" or empty string if format is invalid
func extractBackupUri(backupName string) string {
	parts := strings.Split(backupName, "/")
	if len(parts) < 6 {
		return ""
	}

	locationId := parts[3]
	backupId := parts[5]
	return locationId + "/" + backupId
}

// extractBackupLocation extracts just the location_id from the full backup name
// in format: projects/{project_number}/locations/{location_id}/backups/{backup_id}
// Returns the location_id or empty string if format is invalid
func extractBackupLocation(backupName string) string {
	parts := strings.Split(backupName, "/")
	if len(parts) < 4 {
		return ""
	}

	return parts[3]
}
