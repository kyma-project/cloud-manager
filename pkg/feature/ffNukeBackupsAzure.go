package feature

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
)

var FFNukeBackupsAzure = &nukeBackupsAzure{}

const nukeBackupsAzureFlagName = "nukeBackupsAzure"

type nukeBackupsAzure struct{}

func (k *nukeBackupsAzure) Value(ctx context.Context) bool {

	ffCtx := ContextBuilderFromCtx(ctx).Provider("azure").Feature(types.FeatureNfsBackup).Build(ctx)
	enabled := !ApiDisabled.Value(ffCtx) || provider.BoolVariation(ctx, nukeBackupsAzureFlagName, false)

	return enabled
}

func (k *nukeBackupsAzure) Predicate() composed.Predicate {
	return func(ctx context.Context, _ composed.State) bool {
		return k.Value(ctx)
	}
}
