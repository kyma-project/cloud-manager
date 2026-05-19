package security

import (
	"context"
	"fmt"

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
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error creating data resource group: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error creating security resource group", ctx)
	}

	state.resourceGroupData = rg

	return nil, ctx
}
