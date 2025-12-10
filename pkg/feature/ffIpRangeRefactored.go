package feature

import (
	"context"
)

const ipRangeRefactoredFlagName = "ipRangeRefactored"

var IpRangeRefactored = &ipRangeRefactoredInfo{}

type ipRangeRefactoredInfo struct{}

func (k *ipRangeRefactoredInfo) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, ipRangeRefactoredFlagName, false)
}
