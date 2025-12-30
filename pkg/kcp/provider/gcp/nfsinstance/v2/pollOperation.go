package v2

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
)

// pollOperation checks the status of a pending GCP operation.
func pollOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	opName := nfsInstance.Status.OpIdentifier

	// If no pending operation, continue to next action
	if opName == "" {
		return nil, nil
	}

	logger.Info("Checking GCP Operation Status", "operation", opName)

	// Check operation status
	isDone, err := state.GetFilestoreClient().GetOperation(ctx, opName)
	if err != nil {
		logger.Error(err, "Error getting Filestore Operation from GCP")

		// Clear the operation identifier on error (might be invalid/not found)
		nfsInstance.Status.OpIdentifier = ""

		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: "Error checking operation status",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Operation not completed yet, requeue and wait
	if !isDone {
		logger.Info("Operation still in progress, requeuing")
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime), nil
	}

	// Operation completed, clear the operation identifier
	logger.Info("Operation completed")
	nfsInstance.Status.OpIdentifier = ""

	return composed.UpdateStatus(nfsInstance).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
