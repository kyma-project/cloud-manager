package feature

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

const nukeBackupsAwsFlagName = "nukeBackupsAws"

var FFNukeBackupsAws = &nukeBackupsAwsInfo{}

type nukeBackupsAwsInfo struct{}

func (k *nukeBackupsAwsInfo) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, nukeBackupsAwsFlagName, false)
}

func (k *nukeBackupsAwsInfo) Predicate() composed.Predicate {
	return func(ctx context.Context, _ composed.State) bool {
		return k.Value(ctx)
	}
}
