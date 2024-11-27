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
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	peeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKymaVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.kymaVpcPeering != nil {
		return nil, nil
	}

	//First we need to check if the remote VPC is tagged with the shoot name.
	var isVpcTagged bool
	var err error

	if peeringconfig.VpcPeeringConfig.NetworkTag != "" {
		isVpcTagged, err = state.client.CheckRemoteNetworkTags(ctx, state.remoteNetwork.Spec.Network.Reference.Gcp.NetworkName, state.remoteNetwork.Spec.Network.Reference.Gcp.GcpProject, peeringconfig.VpcPeeringConfig.NetworkTag)

		if err != nil {
			logger.Error(err, "[KCP GCP VPCPeering createKymaVpcPeering] Error creating GCP Kyma VPC Peering while checking any remote network tags")
			return err, ctx
		}
	}

	if !isVpcTagged {
		isVpcTagged, err = state.client.CheckRemoteNetworkTags(ctx, state.remoteNetwork.Spec.Network.Reference.Gcp.NetworkName, state.remoteNetwork.Spec.Network.Reference.Gcp.GcpProject, state.Scope().Spec.ShootName)

		if err != nil {
			logger.Error(err, "[KCP GCP VPCPeering createKymaVpcPeering] Error creating GCP Kyma VPC Peering while checking remote network tags")
			return err, ctx
		}
	}

	if !isVpcTagged {
		logger.Error(err, "[KCP GCP VPCPeering createKymaVpcPeering] Remote network "+state.remoteNetwork.Spec.Network.Reference.Gcp.NetworkName+" is not tagged with the kyma shoot name "+state.Scope().Spec.ShootName)
		state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected
		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: fmt.Sprintf("Error creating VpcPeering, remote VPC does not have a tag with the key: %s", state.Scope().Spec.ShootName),
			}).
			ErrorLogMessage("Error creating Remote VpcPeering").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(5*util.Timing.T60000ms())).
			Run(ctx, state)
	}

	err = state.client.CreateKymaVpcPeering(
		ctx,
		state.getKymaVpcPeeringName(),
		state.remoteNetwork.Spec.Network.Reference.Gcp.NetworkName,
		state.remoteNetwork.Spec.Network.Reference.Gcp.GcpProject,
		state.importCustomRoutes,
		state.localNetwork.Spec.Network.Reference.Gcp.GcpProject,
		state.localNetwork.Spec.Network.Reference.Gcp.NetworkName)

	if err != nil {
		state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected
		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: fmt.Sprintf("Error creating Remote VpcPeering %s", err),
			}).
			ErrorLogMessage("Error creating Remote VpcPeering").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}
	logger.Info("[KCP GCP VPCPeering createKymaVpcPeering] Kyma VPC Peering Connection created")
	return composed.StopWithRequeueDelay(3 * util.Timing.T10000ms()), nil
}
