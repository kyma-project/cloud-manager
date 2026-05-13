package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func storageAccountDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.storageAccount == nil {
		return nil, ctx
	}

	accountName := ptr.Deref(state.storageAccount.Name, "")
	logger.Info("Deleting storage account", "name", accountName)

	_, err := state.azureClient.DeleteStorageAccount(ctx,
		state.resourceGroupDataName(),
		accountName,
		nil)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return azuremeta.LogErrorAndReturn(err, "Error deleting storage account", ctx)
	}

	state.storageAccount = nil

	return nil, ctx
}
