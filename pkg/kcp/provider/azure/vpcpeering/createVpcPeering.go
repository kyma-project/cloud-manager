package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.peering != nil {
		return nil, nil
	}

	resourceGroupName := state.Scope().Spec.Scope.Azure.VpcNetwork // TBD resourceGroup name have the same name as VPC
	virtualNetworkPeeringName := obj.Name

	err := state.client.CreatePeering(ctx,
		resourceGroupName,
		state.Scope().Spec.Scope.Azure.VpcNetwork,
		virtualNetworkPeeringName,
		obj.Spec.VpcPeering.Azure.RemoteVnet,
		true,
	)

	if err != nil {
		logger.Error(err, "Error creating VPC Peering")

		message := azuremeta.GetErrorMessage(err)

		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: message,
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed creating vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	logger.Info("Azure VPC Peering created")

	return composed.StopWithRequeue, ctx
}
