package nuke

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func allDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	someObjectStillExist := false
	for _, rks := range state.Resources {
		if len(rks.Objects) > 0 {
			someObjectStillExist = true
			break
		}
	}

	if !someObjectStillExist {
		logger.Info("All orphan resources nuke deleted")

		state.ObjAsNuke().Status.State = "Completed"

		return composed.PatchStatus(state.ObjAsNuke()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeReady,
				Message: "All orphans deleted",
			}).
			Run(ctx, state)
	}

	logger.Info("Waiting for orphan resources to get nuke deleted")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
