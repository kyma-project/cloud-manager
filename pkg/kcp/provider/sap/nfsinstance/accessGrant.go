package nfsinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func accessGrant(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.accessRight != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Granting SAP share access")

	ar, err := state.sapClient.GrantShareAccess(ctx, state.share.ID, state.Scope().Spec.Scope.OpenStack.Network.Nodes)
	if err != nil {
		logger.Error(err, "Error granting SAP share access")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error granting share access",
			}).
			ErrorLogMessage("Error patching SAP NfsInstance status after grant access error").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.ObjAsNfsInstance().SetStateData(StateDataAccessRightId, ar.ID)

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error patching SAP NfsInstance status with accessRightId").
		SuccessErrorNil().
		Run(ctx, state)
}
