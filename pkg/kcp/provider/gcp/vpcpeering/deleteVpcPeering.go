package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVpcPeering()
	logger := composed.LoggerFromCtx(ctx)

	if state.kymaVpcPeering == nil {
		logger.Info("VPC Peering is not loaded")
		return nil, nil
	}

	logger.Info("Deleting GCP VPC Peering " + obj.Spec.VpcPeering.Gcp.RemotePeeringName)

	err := state.client.DeleteVpcPeering(
		ctx,
		state.getKymaVpcPeeringName(),
		state.Scope().Spec.Scope.Gcp.Project,
		state.Scope().Spec.Scope.Gcp.VpcNetwork,
	)

	if err != nil {
		return err, nil
	}

	return nil, nil
}
