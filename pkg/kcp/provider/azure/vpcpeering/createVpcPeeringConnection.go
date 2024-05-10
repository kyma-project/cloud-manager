package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
)

func createVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	resourceGroupName := state.Scope().Spec.Scope.Azure.VpcNetwork // TBD resourceGroup name have the same name as VPC
	virtualNetworkPeeringName := fmt.Sprintf("%s-%s", obj.Spec.RemoteRef.Namespace, obj.Spec.RemoteRef.Name)

	peering, err := state.client.BeginCreateOrUpdate(ctx,
		resourceGroupName,
		state.Scope().Spec.Scope.Azure.VpcNetwork,
		virtualNetworkPeeringName,
		obj.Spec.VpcPeering.Azure.RemoteVnet,
		obj.Spec.VpcPeering.Azure.AllowVnetAccess,
	)

	if err != nil {
		logger.Error(err, "Error creating VPC Peering")

		return composed.UpdateStatus(obj).
			SetCondition(metav1.Condition{
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

	// TODO should we have different logger values for different providers like connectionId and ID
	logger = logger.WithValues("connectionId", pointer.StringDeref(peering.ID, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("Azure VPC Peering created")

	state.ObjAsVpcPeering().Status.ConnectionId = pointer.StringDeref(peering.ID, "")

	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating VPC Peering status with connection id", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
