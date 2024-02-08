package iprange

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
	name := ipRange.Spec.RemoteRef.Name

	var operation *compute.Operation
	var err error
	switch state.addressOp {
	case client.ADD:
		operation, err = state.computeClient.CreatePscIpRange(ctx, project, vpc, name, name, state.ipAddress, int64(state.prefix))
	case client.MODIFY:
		return composed.UpdateStatus(ipRange).
			SetCondition(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonNotSupported,
				Message: "IpRange update not supported.",
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("IpRange update not supported.").
			Run(ctx, state)
	case client.DELETE:
		operation, err = state.computeClient.DeleteIpRange(ctx, project, name)
	default:
		logger.WithValues("ipRange :", ipRange.Name).Info("Unknown Operation.")
	}

	if err != nil {
		return composed.UpdateStatus(ipRange).
			SetCondition(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			RemoveConditionIfReasonMatched(v1beta1.ConditionTypeError, v1beta1.ReasonGcpError).
			SuccessError(composed.StopWithRequeueDelay(client.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error creating/deleting Address object in GCP :%s", err)).
			Run(ctx, state)
	}
	if operation != nil {
		ipRange.Status.OpIdentifier = operation.Name
		return composed.UpdateStatus(ipRange).
			SuccessError(composed.StopWithRequeueDelay(client.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	return composed.StopWithRequeueDelay(client.GcpOperationWaitTime), nil
}
