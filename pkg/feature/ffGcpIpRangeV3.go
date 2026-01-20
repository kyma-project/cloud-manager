package feature

import (
	"context"
)

const gcpIpRangeV3FlagName = "gcpIpRangeV3"

var GcpIpRangeV3 = &gcpIpRangeV3Info{}

type gcpIpRangeV3Info struct{}

func (k *gcpIpRangeV3Info) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, gcpIpRangeV3FlagName, false)
}
