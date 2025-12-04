/*
required GCP permissions
=========================
  - The service account used to create the VPC peering connection needs the following permissions:
  ** Creates the VPC peering connection
  compute.networks.addPeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/addPeering
  ** Removes the VPC peering connection
  compute.networks.removePeering => https://cloud.google.com/compute/docs/reference/rest/v1/networks/removePeering
  ** Gets the network (VPCs) in order to retrieve the peerings
  compute.networks.get => https://cloud.google.com/compute/docs/reference/rest/v1/networks/get
*/

package vpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createLocalVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) || (state.localVpcPeering != nil || (state.localOperation != nil)) {
		return nil, nil
	}

	op, err := state.client.CreateLocalVpcPeering(
		ctx,
		state.getKymaVpcPeeringName(),
		state.RemoteNetwork().Status.Network.Gcp.NetworkName,
		state.RemoteNetwork().Status.Network.Gcp.GcpProject,
		state.importCustomRoutes,
		state.LocalNetwork().Status.Network.Gcp.GcpProject,
		state.LocalNetwork().Status.Network.Gcp.NetworkName)

	if err != nil {
		state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected
		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: "Error creating local network VpcPeering",
			}).
			ErrorLogMessage("Error creating local network VpcPeering").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.ObjAsVpcPeering().Status.Operation = ptr.Deref(op.Name, "OperationUnknown")
	err = state.PatchObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error updating status with Local VPC Peering operation.",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}
	logger.Info("[KCP GCP VPCPeering createLocalVpcPeering] Local VPC Peering Connection created ", "localVpcPeering", state.getKymaVpcPeeringName())
	return composed.StopWithRequeueDelay(3 * util.Timing.T10000ms()), ctx
}
