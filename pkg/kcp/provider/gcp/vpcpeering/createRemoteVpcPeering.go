/*
required GCP permissions
=========================
  - The service account used to create the VPC peering connection needs the following permissions:
  ** Creates the VPC peering connection
  compute.networks.addPeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/addPeering
  ** Gets the network (VPCs) in order to retrieve the peerings
  compute.networks.get => https://cloud.google.com/compute/docs/reference/rest/v1/networks/get
  ** Fetches the remote network tags
  compute.networks.ListEffectiveTags => https://cloud.google.com/resource-manager/reference/rest/v3/tagKeys/get
*/

package vpcpeering

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createRemoteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.remoteVpcPeering != nil || state.remotePeeringOperation != nil {
		return nil, nil
	}

	op, err := state.client.CreateRemoteVpcPeering(
		ctx,
		state.remotePeeringName,
		state.RemoteNetwork().Status.Network.Gcp.NetworkName,
		state.RemoteNetwork().Status.Network.Gcp.GcpProject,
		state.importCustomRoutes,
		state.LocalNetwork().Status.Network.Gcp.GcpProject,
		state.LocalNetwork().Status.Network.Gcp.NetworkName)

	if err != nil {
		logger.Error(err, "[KCP GCP VpcPeering createRemoteVpcPeering] Error creating Remote VpcPeering")
		state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected

		if gcpmeta.IsNotAuthorized(err) {
			return composed.UpdateStatus(state.ObjAsVpcPeering()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  "True",
					Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
					Message: fmt.Sprintf("Error creating Remote VpcPeering for network %s/%s due to insufficient permissions", state.RemoteNetwork().Status.Network.Gcp.GcpProject, state.RemoteNetwork().Status.Network.Gcp.NetworkName),
				}).
				ErrorLogMessage("Error creating Remote VpcPeering due to insufficient permissions").
				FailedError(composed.StopWithRequeue).
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
				Run(ctx, state)
		}

		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: "Error creating Remote VpcPeering",
			}).
			ErrorLogMessage("Error creating Remote VpcPeering").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.ObjAsVpcPeering().Status.RemotePeeringOperation = ptr.Deref(op.Name, "RemoteOperationUnknown")
	err = state.PatchObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error updating status with Remote VPC Peering operation.",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}
	logger.Info("Remote VPC Peering Connection requested", "operation id", state.ObjAsVpcPeering().Status.RemotePeeringOperation)
	return composed.StopWithRequeueDelay(3 * util.Timing.T10000ms()), ctx
}
