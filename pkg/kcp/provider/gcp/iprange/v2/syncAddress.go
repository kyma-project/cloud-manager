package v2

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/compute/v1"
)

func syncAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Saving GCP Address")

	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork
	name := GetIpRangeName(ipRange.GetName())

	var operation *compute.Operation
	var err error
	switch state.addressOp {
	case client.ADD:
		operation, err = state.computeClient.CreatePscIpRange(ctx, project, vpc, name, "Kyma cloud-manager IP Range", state.ipAddress, int64(state.prefix))
	case client.MODIFY:
		logger.WithValues("ipRange :", ipRange.Name).Info("IpRange update not supported.")
		return composed.StopAndForget, nil
	case client.DELETE:
		operation, err = state.computeClient.DeleteIpRange(ctx, project, state.address.Name)
	default:
		logger.WithValues("ipRange :", ipRange.Name).Info("Unknown Operation.")
	}

	if err != nil {
		return composed.UpdateStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error creating/deleting Address object in GCP :%s", err)).
			Run(ctx, state)
	}
	if operation != nil {
		ipRange.Status.OpIdentifier = operation.Name
		return composed.UpdateStatus(ipRange).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	return composed.StopWithRequeueDelay(state.gcpConfig.GcpOperationWaitTime), nil
}
