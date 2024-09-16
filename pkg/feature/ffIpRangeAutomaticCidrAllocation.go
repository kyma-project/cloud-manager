package feature

import "context"

const ipRangeAutomaticCidrAllocationFlagName = "ipRangeAutomaticCidrAllocation"

// Deprecated: Do not use anymore
var IpRangeAutomaticCidrAllocation = &ipRangeAutomaticCidrAllocationInfo{}

type ipRangeAutomaticCidrAllocationInfo struct{}

func (k *ipRangeAutomaticCidrAllocationInfo) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, ipRangeAutomaticCidrAllocationFlagName, true)
}
