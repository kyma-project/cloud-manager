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
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
)

func createRemoteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.remoteVpcPeering != nil {
		return nil, nil
	}

	err := state.client.CreateRemoteVpcPeering(
		ctx,
		state.remotePeeringName,
		state.remoteNetwork.Spec.Network.Reference.Gcp.NetworkName,
		state.remoteNetwork.Spec.Network.Reference.Gcp.GcpProject,
		state.importCustomRoutes,
		state.localNetwork.Spec.Network.Reference.Gcp.GcpProject,
		state.localNetwork.Spec.Network.Reference.Gcp.NetworkName)

	if err != nil {
		message := fmt.Sprintf("Error creating Remote VpcPeering %s", err)
		// If we already have a peering with the same network and project, we need to let the user know that the peering already exists
		// and he might need to either delete the existing peering or use the same name for the new peering. This is required since we don't
		// delete any objects on the user project.
		matchesExistingPeering, regexError := regexp.Match("There is already a peering (.*) with the same network. Select another network.", []byte(err.Error()))
		if regexError != nil {
			return err, nil
		}
		if matchesExistingPeering {
			message = fmt.Sprintf("Error creating Remote VpcPeering: %s Please check the VPC peerings on your project.", err)
		}
		state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected
		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: message,
			}).
			ErrorLogMessage("Error creating Remote VpcPeering").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}
	logger.Info("Remote VPC Peering Connection created")
	return composed.StopWithRequeueDelay(3 * util.Timing.T10000ms()), nil
}
