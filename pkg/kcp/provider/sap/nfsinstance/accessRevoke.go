package nfsinstance

import (
	"context"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func accessRevoke(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.accessRight == nil {
		return nil, nil
	}

	logger.Info("Revoking SAP NfsInstance access rights")

	err := state.sapClient.RevokeShareAccess(ctx, state.share.ID, state.accessRight.ID)
	if err != nil && !gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error revoking SAP share access",
			}).
			FailedError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			ErrorLogMessage("Error patching SAP NfsInstance status after error revoking access rights").
			Run(ctx, state)
	}

	state.ObjAsNfsInstance().SetStateData(StateDataAccessRightId, "")

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error patching SAP NfsInstance status after access right revoke").
		SuccessErrorNil().
		Run(ctx, state)
}
