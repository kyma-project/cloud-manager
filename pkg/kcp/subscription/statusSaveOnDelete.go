package subscription

import (
	"context"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusSaveOnDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if len(state.resources) == 0 {
		return nil, ctx
	}

	var sb strings.Builder
	firstGvk := true
	for gvk, arr := range state.resources {
		if firstGvk {
			firstGvk = false
		} else {
			sb.WriteString(", ")
		}
		sb.WriteString(gvk.Group)
		sb.WriteString("/")
		sb.WriteString(gvk.Version)
		sb.WriteString("/")
		sb.WriteString(gvk.Kind)
		sb.WriteString(": ")
		firstObj := true
		for _, obj := range arr {
			if firstObj {
				firstObj = false
			} else {
				sb.WriteString(", ")
			}
			sb.WriteString(obj.Name)
		}
	}

	usedBy := sb.String()

	logger := composed.LoggerFromCtx(ctx)
	logger = logger.WithValues("subscriptionUsedBy", usedBy)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.ObjAsSubscription().Status.State = cloudcontrolv1beta1.StateWarning

	return composed.PatchStatus(state.ObjAsSubscription()).
		SetExclusiveConditions(metav1.Condition{
			Type:               cloudcontrolv1beta1.ConditionTypeWarning,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: state.ObjAsSubscription().Generation,
			Reason:             cloudcontrolv1beta1.ReasonDeleteWhileUsed,
			Message:            "Used by: " + usedBy,
		}).
		SuccessLogMsg("Subscription being deleted while still used").
		ErrorLogMessage("Error patching subscription status with warning state while still used").
		SuccessError(composed.StopWithRequeue).
		FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
		Run(ctx, state)

}
