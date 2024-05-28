package feature

import "context"

const ipRangeAutomaticCidrAllocationFlagName = "ipRangeAutomaticCidrAllocation"

var IpRangeAutomaticCidrAllocation = &ipRangeAutomaticCidrAllocationInfo{}

type ipRangeAutomaticCidrAllocationInfo struct{}

func (k *ipRangeAutomaticCidrAllocationInfo) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, ipRangeAutomaticCidrAllocationFlagName, false)
}
