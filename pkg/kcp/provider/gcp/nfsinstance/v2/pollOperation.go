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

	if opName == "" {
		return nil, ctx
	}

	logger.Info("Checking GCP Operation Status", "operation", opName)

	isDone, err := state.GetFilestoreClient().GetOperation(ctx, opName)
	if err != nil {
		logger.Error(err, "Error getting Filestore Operation from GCP")

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

	if !isDone {
		logger.Info("Operation still in progress, requeuing")
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime), nil
	}

	logger.Info("Operation completed")
	nfsInstance.Status.OpIdentifier = ""

	return composed.UpdateStatus(nfsInstance).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
