package vpcpeering

import (
	"cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
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

	// If the remote VPC Peering is still active, we need to wait before deleting it.
	// Otherwise, Google will return a 400 error saying, 'There is a peering operation in progress on the local or peer network. Try again later.'
	// If it's not active, that means the kyma VPC Peering is already deleted, and we can proceed with the deletion of the remote VPC Peering.
	if ptr.Deref(state.remoteVpcPeering.State, "") == computepb.NetworkPeering_ACTIVE.String() {
		logger.Info("Remote VPC Peering is still active. It should wait for the local VPC Peering to be deleted before proceeding with the deletion of the remote VPC Peering.")
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	logger.Info("Deleting Remote GCP VPC Peering " + obj.Spec.Details.PeeringName)

	err := state.client.DeleteVpcPeering(
		ctx,
		obj.Spec.Details.PeeringName,
		state.RemoteNetwork().Status.Network.Gcp.GcpProject,
		state.RemoteNetwork().Status.Network.Gcp.NetworkName,
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
