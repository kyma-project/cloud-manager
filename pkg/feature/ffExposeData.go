package feature

import (
	"context"
)

const exposeDataFlagName = "exposeData"

var ExposeData = &exposeDataInfo{}

type exposeDataInfo struct{}

func (f *exposeDataInfo) Value(ctx context.Context) bool {
	v := provider.BoolVariation(ctx, exposeDataFlagName, true)
	return v
}
