package errorhandling

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HandleError handles the AWS API error, sets the object status conditions if it implements
// composed.ObjWithConditions, sets the object status state if it implements composed.ObjWithConditionsAndState,
// calls composed.PatchStatus() to patch the object status and returns error that composed.Action should return,
// all as explained in https://github.tools.sap/I517584/kyma-docs/blob/main/concepts/handling-cloud-api-errors/README.md
func HandleError(
	ctx context.Context,
	err error,
	state composed.State,
	description string,
	conditionReason string,
	conditionMessage string,
) error {
	if err == nil {
		return nil
	}
	logger := composed.LoggerFromCtx(ctx)
	logger = logger.
		WithValues(
			"error", err.Error(),
		)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	if awsmeta.IsErrorRetryable(err) {
		logger.Info("AWS Retryable Error: " + description)
		return awsmeta.ErrorToRequeueResponse(err)
	}

	if os, ok := state.Obj().(composed.ObjWithConditionsAndState); ok {
		os.SetState("Error")
	}
	if os, ok := state.Obj().(composed.ObjWithConditions); ok {
		ctx = composed.LoggerIntoCtx(ctx, logger.WithCallDepth(1))
		res, _ := composed.PatchStatus(os).
			SetExclusiveConditions(metav1.Condition{
				Type:    "Error",
				Status:  metav1.ConditionTrue,
				Reason:  conditionReason,
				Message: conditionMessage,
			}).
			ErrorLogMessage("Error patching status: "+description).
			SuccessLogMsg("Long delaying: "+description).
			SuccessError(awsmeta.ErrorToRequeueResponse(err)).
			Run(ctx, state)
		return res
	}
	return awsmeta.ErrorToRequeueResponse(err)
}
