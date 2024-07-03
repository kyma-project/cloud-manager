package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVpcPeering()
	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Deleting GCP VPC Peering " + obj.Spec.VpcPeering.Gcp.PeeringName)

	_, err := state.client.DeleteVpcPeering(
		ctx,
		&obj.Spec.VpcPeering.Gcp.PeeringName,
		&state.Scope().Spec.Scope.Gcp.Project,
		&state.Scope().Spec.Scope.Gcp.VpcNetwork,
	)

	if err != nil {
		return err, ctx
	}

	return nil, nil
}
