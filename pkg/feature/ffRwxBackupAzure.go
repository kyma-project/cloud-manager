package feature

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

var FFRwxBackupAzure = &rwxBackupAzure{}

const rwxBackupAzureFlagName = "rwxBackupAzure"

type rwxBackupAzure struct{}

func (k *rwxBackupAzure) Value(ctx context.Context) bool {
	ffCtxAttr := MustContextFromCtx(ctx).GetCustom()
	isAzure := strings.Contains(
		util.CastInterfaceToString(ffCtxAttr[types.KeyProvider]), "azure")
	isRwxBackup := strings.Contains(
		util.CastInterfaceToString(ffCtxAttr[types.KeyFeature]), types.FeatureNfsBackup)
	return (isAzure && isRwxBackup) || provider.BoolVariation(ctx, rwxBackupAzureFlagName, false)
}

func (k *rwxBackupAzure) Predicate() composed.Predicate {
	return func(ctx context.Context, _ composed.State) bool {
		return k.Value(ctx)
	}
}
