package subscription

import (
	"context"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func labelBindingName(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsSubscription().Spec.Details.Garden == nil {
		return nil, ctx
	}

	expectedBindingName := state.ObjAsSubscription().Spec.Details.Garden.BindingName
	expectedProvider := strings.ToLower(string(state.provider))

	labelsToSet := map[string]string{}

	val, hasLabel := state.ObjAsSubscription().Labels[cloudcontrolv1beta1.SubscriptionLabelBindingName]
	if !hasLabel || expectedBindingName != val {
		labelsToSet[cloudcontrolv1beta1.SubscriptionLabelBindingName] = expectedBindingName
	}

	val, hasLabel = state.ObjAsSubscription().Labels[cloudcontrolv1beta1.SubscriptionLabelProvider]
	if !hasLabel || expectedProvider != val {
		labelsToSet[cloudcontrolv1beta1.SubscriptionLabelProvider] = expectedProvider
	}

	if len(labelsToSet) == 0 {
		return nil, ctx
	}

	_, err := composed.PatchObjMergeLabels(ctx, state.ObjAsSubscription(), state.Cluster().K8sClient(), labelsToSet)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching subscription with label for binding name", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
	}

	return nil, ctx
}
