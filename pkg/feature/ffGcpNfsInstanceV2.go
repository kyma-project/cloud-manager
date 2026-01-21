package feature

import (
	"context"
)

const GcpNfsInstanceV2FlagName = "gcpNfsInstanceV2"

var GcpNfsInstanceV2 = &gcpNfsInstanceV2Info{}

type gcpNfsInstanceV2Info struct{}

func (k *gcpNfsInstanceV2Info) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, GcpNfsInstanceV2FlagName, false)
}
