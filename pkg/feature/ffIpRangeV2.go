package feature

import (
	"context"
)

const ipRangeV2FlagName = "ipRangeV2"

var IpRangeV2 = &ipRangeV2Info{}

type ipRangeV2Info struct{}

func (k *ipRangeV2Info) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, ipRangeV2FlagName, false)
}
