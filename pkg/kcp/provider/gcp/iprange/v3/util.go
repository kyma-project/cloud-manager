package v3

import "fmt"

// GetIpRangeName generates the GCP resource name for an IpRange from the K8s resource name.
// Format: cm-<uuid>
// This is the new naming convention (v2).
func GetIpRangeName(kcpResourceName string) string {
	return fmt.Sprintf("cm-%s", kcpResourceName)
}
