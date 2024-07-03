package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func createVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeeringConnection != nil {
		return nil, nil
	}

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	con, err := state.client.CreateVpcPeering(
		ctx,
		state.peeringName,
		state.remoteVpc,
		state.remoteProject,
		state.importCustomRoutes,
		&project,
		&vpc)

	if err != nil {
		logger.Error(err, "Error creating VPC Peering")

		if err.Error() == "remote network "+*state.remoteVpc+" is not tagged with the kyma shoot name "+vpc {
			return composed.UpdateStatus(state.ObjAsVpcPeering()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  "True",
					Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork, //I believe we should change it for something like ReasonRemoteNetworkNotTagged
					Message: fmt.Sprintf("Remote network %s is not tagged with the kyma shoot name %s", *state.remoteVpc, vpc),
				}).
				ErrorLogMessage("Remote network is not tagged with the kyma shoot name").
				FailedError(composed.StopWithRequeue).
				SuccessError(composed.StopWithRequeueDelay(time.Minute)).
				Run(ctx, state)
		}

		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: fmt.Sprintf("Failed creating VpcPeerings %s", err),
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed creating vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(time.Minute)).
			Run(ctx, state)
	}

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("GCP VPC Peering Connection created")

	state.vpcPeeringConnection = con

	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating VPC Peering status", composed.StopWithRequeue, ctx)
	}
	return nil, ctx
}
