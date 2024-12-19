package v2

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkGcpOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	opName := ipRange.Status.OpIdentifier
	logger.WithValues("ipRange", ipRange.Name).Info("Checking Operation")

	//If no OpIdentifier, then continue to next action.
	if opName == "" {
		return nil, nil
	}

	//Check SyncPsa Operation..
	if ipRange.Status.State == client.SyncPsaConnection ||
		ipRange.Status.State == client.DeletePsaConnection {
		op, err := state.serviceNetworkingClient.GetServiceNetworkingOperation(ctx, opName)
		if err != nil {

			//If the operation is not found, reset the OpIdentifier.
			var e *googleapi.Error
			if ok := errors.As(err, &e); ok {
				if e.Code == 404 {
					ipRange.Status.OpIdentifier = ""
				}
			}

			return composed.PatchStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: err.Error(),
				}).
				SuccessError(composed.StopWithRequeue).
				SuccessLogMsg("Error getting Service Networking Operation from GCP.").
				Run(ctx, state)
		}

		//Operation not completed yet.. requeue again.
		if op != nil && !op.Done {
			return composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), nil
		}

		//If not able to find the operation or it is completed, reset OpIdentifier.
		if op == nil || op.Done {
			ipRange.Status.OpIdentifier = ""
		}

		//If the operation failed, update the error status on the object.
		if op != nil && op.Error != nil {
			return composed.PatchStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: op.Error.Message,
				}).
				SuccessError(composed.StopWithRequeue).
				SuccessLogMsg(fmt.Sprintf("Service Networking Operation error : %s", op.Error.Message)).
				Run(ctx, state)
		}
	} else if ipRange.Status.State == client.SyncAddress ||
		ipRange.Status.State == client.DeleteAddress {
		project := state.Scope().Spec.Scope.Gcp.Project
		op, err := state.computeClient.GetGlobalOperation(ctx, project, opName)
		if err != nil {
			return composed.PatchStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: err.Error(),
				}).
				SuccessError(composed.StopWithRequeue).
				SuccessLogMsg("Error getting Compute Operation from GCP.").
				Run(ctx, state)
		}

		//Operation not completed yet.. requeue again.
		if op != nil && op.Status != "DONE" {
			return composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), nil
		}

		//If not able to find the operation or it is completed, reset OpIdentifier.
		if op == nil || op.Status == "DONE" {
			ipRange.Status.OpIdentifier = ""
		}

		//If the operation failed, update the error status on the object.
		if op != nil && op.Error != nil {
			msg := op.StatusMessage
			if msg == "" {
				if len(op.Error.Errors) > 0 {
					msg = op.Error.Errors[0].Message
				} else {
					msg = "Operation failed with no error message."
				}
			}
			return composed.PatchStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: msg,
				}).
				SuccessError(composed.StopWithRequeue).
				SuccessLogMsg(fmt.Sprintf("Compute Operation error : %s", op.StatusMessage)).
				Run(ctx, state)
		}

	}

	return nil, nil
}
