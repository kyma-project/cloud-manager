package security

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func resourceGroupDataDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroupData == nil {
		return nil, ctx
	}

	logger.Info("Deleting security resource group")
	_, err := azureclient.PollUntilDone(state.azureClient.DeleteResourceGroup(
		ctx,
		ResourceGroupDataName(state.ObjAsRuntime().Spec.Shoot.Name),
	))(ctx, nil)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error deleting data resource group: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error deleting security resource group", ctx)
	}

	state.resourceGroupData = nil

	return nil, ctx
}
