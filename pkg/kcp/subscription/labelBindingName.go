package subscription

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func labelBindingName(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsSubscription().Spec.Details.Garden == nil {
		return nil, ctx
	}

	bindingName := state.ObjAsSubscription().Spec.Details.Garden.BindingName

	val, hasLabel := state.ObjAsSubscription().Labels[cloudcontrolv1beta1.LabelSubscriptionBindingName]
	if hasLabel && bindingName == val {
		return nil, ctx
	}

	_, err := composed.PatchObjMergeLabel(ctx, state.ObjAsSubscription(), state.Cluster().K8sClient(), cloudcontrolv1beta1.LabelSubscriptionBindingName, bindingName)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching subscription with label for binding name", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
	}

	return nil, ctx
}
