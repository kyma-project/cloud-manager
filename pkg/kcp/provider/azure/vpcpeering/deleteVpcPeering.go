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

	if len(obj.Status.Id) == 0 {
		logger.Info("VpcPeering deleted before Azure peering is created")
		return nil, nil
	}

	resourceGroupName := state.Scope().Spec.Scope.Azure.VpcNetwork

	logger.Info("Deleting VpcPeering")

	err := state.localClient.DeletePeering(
		ctx,
		resourceGroupName,
		state.Scope().Spec.Scope.Azure.VpcNetwork,
		obj.GetLocalPeeringName(),
	)

	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error deleting vpc peering", ctx)
	}

	logger.Info("VpcPeering deleted")

	return nil, nil
}
