package feature

import "context"

const gcpNfsVolumeAutomaticLocationAllocationFlagName = "gcpNfsVolumeAutomaticLocationAllocation"

var GcpNfsVolumeAutomaticLocationAllocation = &gcpNfsVolumeAutomaticLocationAllocationInfo{}

type gcpNfsVolumeAutomaticLocationAllocationInfo struct{}

func (k *gcpNfsVolumeAutomaticLocationAllocationInfo) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, gcpNfsVolumeAutomaticLocationAllocationFlagName, true)
}
