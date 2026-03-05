package v2

import "fmt"

// GetFilestoreParentPath returns the parent path for Filestore resources.
// Format: projects/{project}/locations/{location}
// Use "-" for location to list across all locations.
func GetFilestoreParentPath(projectId, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", projectId, location)
}

// GetFilestoreInstancePath returns the full GCP Filestore instance resource name.
// Format: projects/{project}/locations/{location}/instances/{name}
func GetFilestoreInstancePath(projectId, location, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectId, location, name)
}
