package feature

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var FFRwxBackupAzure = &rwxBackupAzure{}

const rwxBackupAzureFlagName = "rwxBackupAzure"

type rwxBackupAzure struct{}

func (k *rwxBackupAzure) Value(ctx context.Context) bool {

	ffCtx := ContextBuilderFromCtx(ctx).Provider("azure").Feature(types.FeatureNfsBackup).Build(ctx)
	enabled := !ApiDisabled.Value(ffCtx) || provider.BoolVariation(ctx, rwxBackupAzureFlagName, false)

	log.FromContext(ctx).Info(fmt.Sprintf("Azure RWX Backup feature flag : %v", enabled))
	return enabled
}

func (k *rwxBackupAzure) Predicate() composed.Predicate {
	return func(ctx context.Context, _ composed.State) bool {
		return k.Value(ctx)
	}
}
