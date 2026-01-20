package client

import "fmt"

// GetFilestoreInstanceId returns the GCP Filestore instance ID with "cm-" prefix.
// GCP instance IDs must be 1-63 characters. If the resulting name exceeds this,
// GCP will return an error rather than silently truncating.
// This follows the same pattern as GetGcpMemoryStoreRedisInstanceId.
func GetFilestoreInstanceId(kcpResourceName string) string {
	return fmt.Sprintf("cm-%s", kcpResourceName)
}

// GetFilestoreName returns the full GCP Filestore resource name.
// Format: projects/{project}/locations/{location}/instances/cm-{instanceId}
// This follows the same pattern as GetGcpMemoryStoreRedisName.
func GetFilestoreName(projectId, location, kcpResourceName string) string {
	return fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, location, GetFilestoreInstanceId(kcpResourceName))
}
