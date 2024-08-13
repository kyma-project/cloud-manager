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
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKymaVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.kymaVpcPeering != nil {
		return nil, nil
	}

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	//First we need to check if the remote VPC is tagged with the shoot name.
	isVpcTagged, err := state.client.CheckRemoteNetworkTags(ctx, state.remoteVpc, state.remoteProject, state.Scope().Spec.ShootName)
	if err != nil {
		logger.Error(err, "Error creating GCP Kyma VPC Peering while checking remote network tags")
		return err, ctx
	}

	if !isVpcTagged {
		logger.Error(err, "Remote network "+state.remoteVpc+" is not tagged with the kyma shoot name "+state.Scope().Spec.ShootName)
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
		state.remoteVpc,
		state.remoteProject,
		state.importCustomRoutes,
		project,
		vpc)

	if err != nil {
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
	logger.Info("Kyma VPC Peering Connection created")
	return composed.StopWithRequeueDelay(3 * util.Timing.T10000ms()), nil
}
