package vpcpeering

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadRemoteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.remoteVpcPeering != nil {
		return nil, nil
	}

	logger.Info("[KCP GCP VpcPeering loadRemoteVpcPeering] Loading Remote VPC Peering")

	remoteVpcPeering, err := state.client.GetVpcPeering(ctx, state.remotePeeringName, state.RemoteNetwork().Status.Network.Gcp.GcpProject, state.RemoteNetwork().Status.Network.Gcp.NetworkName)
	if err != nil {
		if composed.IsMarkedForDeletion(state.ObjAsVpcPeering()) {
			logger.Info("[KCP GCP VPCPeering loadRemoteVPCPeering] GCP KCP VpcPeering Error fetching Remote Network, proceeding with deletion")
			return nil, nil
		}
		if gcpmeta.IsNotFound(err) { // Returns not found when network is not found (not the peering) Since it is only possible to fetch the peerings directly from the network
			logger.Error(err, "[KCP GCP VpcPeering loadRemoteVpcPeering] Remote VPC Network not found")
			state.ObjAsVpcPeering().Status.State = v1beta1.VirtualNetworkPeeringStateDisconnected
			meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  v1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: fmt.Sprintf("Remote VPC Network %s/%s not found.", state.RemoteNetwork().Status.Network.Gcp.GcpProject, state.RemoteNetwork().Status.Network.Gcp.NetworkName),
			})
			err = state.UpdateObjStatus(ctx)
			if err != nil {
				return composed.LogErrorAndReturn(err,
					"Error updating status, the remote VPC network does not exist.",
					composed.StopAndForget,
					ctx,
				)
			}
		}
		if gcpmeta.IsNotAuthorized(err) {
			logger.Error(err, "[KCP GCP VpcPeering loadRemoteVpcPeering] Could not fetch Remote VPC Network, required permissions are missing")
			state.ObjAsVpcPeering().Status.State = v1beta1.VirtualNetworkPeeringStateDisconnected
			meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  v1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: fmt.Sprintf("Unable to fetch remote VPC network %s/%s due to insufficient permissions.", state.RemoteNetwork().Status.Network.Gcp.GcpProject, state.RemoteNetwork().Status.Network.Gcp.NetworkName),
			})
			err = state.UpdateObjStatus(ctx)
			if err != nil {
				return composed.LogErrorAndReturn(err,
					"Error updating status, no permissions required to fetch remote VPC network.",
					composed.StopAndForget,
					ctx,
				)
			}
		}

		state.ObjAsVpcPeering().Status.State = v1beta1.VirtualNetworkPeeringStateDisconnected
		logger.Error(err, "[KCP GCP VpcPeering loadRemoteVpcPeering] Error loading Remote VpcPeering")
		meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: "Error loading Remote Vpc Peering",
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating status since it was not possible to load the remote Vpc Peering",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	logger.Info("[KCP GCP VpcPeering createRemoteVpcPeering] Remote VPC Peering loaded")
	state.remoteVpcPeering = remoteVpcPeering
	return nil, nil
}
