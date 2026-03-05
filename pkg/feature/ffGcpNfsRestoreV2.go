package feature

import (
	"context"
)

const GcpNfsRestoreV2FlagName = "gcpNfsRestoreV2"

var GcpNfsRestoreV2 = &gcpNfsRestoreV2Info{}

type gcpNfsRestoreV2Info struct{}

func (k *gcpNfsRestoreV2Info) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, GcpNfsRestoreV2FlagName, false)
}
