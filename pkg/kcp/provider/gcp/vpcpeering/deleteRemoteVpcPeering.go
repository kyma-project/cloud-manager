package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteRemoteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVpcPeering()
	logger := composed.LoggerFromCtx(ctx)

	if state.remoteVpcPeering == nil {
		logger.Info("VPC Peering is not loaded")
		return nil, nil
	}

	if !obj.Spec.Details.DeleteRemotePeering {
		logger.Info("DeleteRemotePeering is set to false. Skipping deletion of remote VPC Peering")
		return nil, nil
	}

	logger.Info("Deleting Remote GCP VPC Peering " + obj.Spec.Details.PeeringName)

	err := state.client.DeleteVpcPeering(
		ctx,
		obj.Spec.Details.PeeringName,
		state.remoteNetwork.Status.Network.Gcp.GcpProject,
		state.remoteNetwork.Status.Network.Gcp.NetworkName,
	)

	if err != nil {
		if gcpmeta.IsNotAuthorized(err) {
			logger.Info("Not authorized to delete remote VPC Peering")
		}
		if gcpmeta.IsTooManyRequests(err) {
			logger.Info("Too many requests. Requeueing")
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
		}
		logger.Error(err, "Error deleting remote VPC Peering")
	}

	return nil, nil
}
