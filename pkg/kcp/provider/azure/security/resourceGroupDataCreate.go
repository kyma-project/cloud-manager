package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func resourceGroupDataCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroupData != nil {
		return nil, ctx
	}

	logger.Info("Creating security resource group")
	tags := map[string]string{
		tagKymaRuntimeId: state.ObjAsRuntime().Name,
		tagKymaShootName: state.shootName(),
	}
	rg, err := state.azureClient.CreateResourceGroup(ctx,
		state.resourceGroupDataName(),
		state.location(),
		tags)
	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error creating security resource group", ctx)
	}

	state.resourceGroupData = rg
	return nil, ctx
}
