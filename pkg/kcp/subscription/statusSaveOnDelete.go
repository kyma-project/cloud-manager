package subscription

import (
	"context"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func statusSaveOnDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if len(state.resources) == 0 {
		sp := composed.NewStatusPatcherComposed(state.ObjAsSubscription())
		state.ObjAsSubscription().RemoveStatusDeleteWhileUsed()
		return sp.
			OnSuccess(composed.Continue).
			OnFailure(composed.Log("Failed to patch Subscription status to remove delete while used condition")).
			Run(ctx, state.Cluster().K8sClient())
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

	return composed.NewStatusPatcherComposed(state.ObjAsSubscription()).
		MutateStatus(func(obj *cloudcontrolv1beta1.Subscription) {
			obj.SetStatusDeleteWhileUsed("Used by: " + usedBy)
		}).
		OnStatusChanged(composed.Log("Subscription being deleted while still used")).
		OnSuccess(composed.RequeueAfter(util.Timing.T10000ms())).
		OnFailure(composed.Log("Error patching subscription status with warning state while still used")).
		Run(ctx, state.Cluster().K8sClient())
}
