package mock

import (
	"context"

	"google.golang.org/api/compute/v1"
)

// IpRangeClientUtils provides test utilities for IpRange mock in Discovery API format (for v2 legacy tests).
type IpRangeClientUtils interface {
	// GetIpRangeDiscovery retrieves an address in Discovery API format (for test assertions)
	GetIpRangeDiscovery(ctx context.Context, projectId, name string) (*compute.Address, error)
	// ListGlobalAddressesDiscovery lists addresses in Discovery API format (for test assertions)
	ListGlobalAddressesDiscovery(ctx context.Context, projectId, vpc string) ([]*compute.Address, error)
}
