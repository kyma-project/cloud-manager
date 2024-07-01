package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	lll := logger.WithValues("vpcPeeringName", obj.Name)

	if len(obj.Status.Id) == 0 {
		lll.Info("VpcPeering deleted before Azure peering is created")
		return nil, nil
	}

	resourceGroupName := state.Scope().Spec.Scope.Azure.VpcNetwork

	lll = lll.WithValues("vpcPeeringId", obj.Status.Id)
	lll.Info("Deleting VpcPeering")

	err := state.client.DeletePeering(
		ctx,
		resourceGroupName,
		state.Scope().Spec.Scope.Azure.VpcNetwork,
		obj.Name,
	)

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error deleting vpc peering", composed.LoggerIntoCtx(ctx, lll))
	}

	return nil, nil
}
