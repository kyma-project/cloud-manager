package iprange

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// waitOperationDone polls GCP to check if an async operation has completed.
// IpRange operations can be either:
// - Compute operations (for address create/delete)
// - Service Networking operations (for PSA connection create/update/delete)
//
// The operation type is determined by the status.state field.
// If the operation is still running, this action requeues with a delay.
// If the operation fails, it sets an error condition.
// If the operation succeeds or doesn't exist, it clears the OpIdentifier.
func waitOperationDone(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	opName := ipRange.Status.OpIdentifier
	logger.WithValues("ipRange", ipRange.Name).Info("Checking Operation")

	// If no OpIdentifier, then continue to next action
	if opName == "" {
		return nil, nil
	}

	// Check operation based on current state
	switch ipRange.Status.State {
	case gcpclient.SyncPsaConnection, gcpclient.DeletePsaConnection:
		return checkServiceNetworkingOperation(ctx, state, opName)
	case gcpclient.SyncAddress, gcpclient.DeleteAddress:
		return checkComputeOperation(ctx, state, opName)
	default:
		// Unknown state, clear operation identifier
		ipRange.Status.OpIdentifier = ""
		return nil, nil
	}
}

// checkServiceNetworkingOperation checks the status of a Service Networking operation.
func checkServiceNetworkingOperation(ctx context.Context, state *State, opName string) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	ipRange := state.ObjAsIpRange()

	op, err := state.serviceNetworkingClient.GetServiceNetworkingOperation(ctx, opName)
	if err != nil {
		// If the operation is not found, reset the OpIdentifier
		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 {
				ipRange.Status.OpIdentifier = ""
				return nil, nil
			}
		}
		logger.Error(err, "Error getting Service Networking Operation from GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Operation not completed yet, requeue with delay
	if op != nil && !op.Done {
		logger.Info("Service Networking Operation is still running", "operation", opName)
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}

	// Operation completed or not found, reset OpIdentifier
	if op == nil || op.Done {
		ipRange.Status.OpIdentifier = ""
	}

	// If the operation failed, update the error status
	if op != nil && op.Error != nil {
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: op.Error.Message,
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg(fmt.Sprintf("Service Networking Operation error: %s", op.Error.Message)).
			Run(ctx, state)
	}

	return nil, nil
}

// checkComputeOperation checks the status of a Compute operation.
func checkComputeOperation(ctx context.Context, state *State, opName string) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	ipRange := state.ObjAsIpRange()

	project := state.Scope().Spec.Scope.Gcp.Project
	op, err := state.computeClient.GetGlobalOperation(ctx, project, opName)
	if err != nil {
		logger.Error(err, "Error getting Compute Operation from GCP")
		return composed.PatchStatus(ipRange).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Operation not completed yet, requeue with delay
	if op != nil && op.Status != nil && *op.Status != computepb.Operation_DONE {
		logger.Info("Compute Operation is still running", "operation", opName, "status", op.Status.String())
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}

	// Operation completed or not found, reset OpIdentifier
	if op == nil || (op.Status != nil && *op.Status == computepb.Operation_DONE) {
		ipRange.Status.OpIdentifier = ""
	}

	// If the operation failed, update the error status
	if op != nil && op.Error != nil {
		msg := ""
		if op.StatusMessage != nil {
			msg = *op.StatusMessage
		}
		if msg == "" {
			if len(op.Error.Errors) > 0 && op.Error.Errors[0].Message != nil {
				msg = *op.Error.Errors[0].Message
			} else {
				msg = "Operation failed with no error message"
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
			SuccessLogMsg(fmt.Sprintf("Compute Operation error: %s", msg)).
			Run(ctx, state)
	}

	return nil, nil
}
