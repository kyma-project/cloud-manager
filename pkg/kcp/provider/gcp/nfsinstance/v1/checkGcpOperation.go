package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"

	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkGcpOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	opName := nfsInstance.Status.OpIdentifier
	logger.WithValues("NfsInstance", nfsInstance.Name).Info("Checking GCP Operation Status")

	//If no OpIdentifier, then continue to next action.
	if opName == "" {
		return nil, nil
	}

	project := state.Scope().Spec.Scope.Gcp.Project
	op, err := state.filestoreClient.GetFilestoreOperation(ctx, project, opName)
	if err != nil {

		//If the operation is not found, reset the OpIdentifier.
		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 {
				nfsInstance.Status.OpIdentifier = ""
			}
		}

		logger.Error(err, "Error getting File Operation from GCP.")
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	//Operation not completed yet.. requeue again.
	if op != nil && !op.Done {
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}

	//If not able to find the operation or it is completed, reset OpIdentifier.
	if op == nil || op.Done {
		nfsInstance.Status.OpIdentifier = ""
	}

	//If the operation failed, update the error status on the object.
	if op != nil && op.Error != nil {
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonGcpError,
				Message: op.Error.Message,
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg(fmt.Sprintf("File Operation error : %s", op.Error.Message)).
			Run(ctx, state)
	}

	return nil, nil
}
