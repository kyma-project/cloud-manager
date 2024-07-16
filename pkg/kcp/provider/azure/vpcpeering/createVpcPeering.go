package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"time"
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

	peering, err := state.client.CreatePeering(ctx,
		resourceGroupName,
		state.Scope().Spec.Scope.Azure.VpcNetwork,
		virtualNetworkPeeringName,
		obj.Spec.VpcPeering.Azure.RemoteVnet,
		obj.Spec.VpcPeering.Azure.AllowVnetAccess,
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
			SuccessError(composed.StopWithRequeueDelay(time.Minute)).
			Run(ctx, state)
	}

	logger = logger.WithValues("id", ptr.Deref(peering.ID, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("Azure VPC Peering created")

	obj.Status.Id = ptr.Deref(peering.ID, "")

	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating VPC Peering status with connection id", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
