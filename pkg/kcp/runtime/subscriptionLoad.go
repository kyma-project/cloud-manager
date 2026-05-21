package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func subscriptionLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	subscription := &cloudcontrolv1beta1.Subscription{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      state.ObjAsRuntime().Spec.Shoot.SecretBindingName,
	}, subscription)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Failed to get subscription by name by binding name", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
	}
	if err != nil {
		// direct mapping bindingName => subscription name failed
		// try to find subscription by label search

		list := &cloudcontrolv1beta1.SubscriptionList{}
		err = state.Cluster().K8sClient().List(ctx, list, client.MatchingLabels{
			cloudcontrolv1beta1.SubscriptionLabelBindingName: state.ObjAsRuntime().Spec.Shoot.SecretBindingName,
		})
		if err != nil {
			return composed.LogErrorAndReturn(err, "Failed to list subscriptions by label binding name", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
		}

		if len(list.Items) == 0 {
			subscription = nil
		} else {
			if len(list.Items) > 1 {
				composed.LoggerFromCtx(ctx).
					WithValues("bindingName", state.ObjAsRuntime().Spec.Shoot.SecretBindingName).
					Info("Multiple subscriptions found matching by label to binding")
			}
			subscription = &list.Items[0]
		}
	}

	state.subscription = subscription

	ctx = composed.LoggerIntoCtx(ctx, composed.LoggerFromCtx(ctx).WithValues("subscription", state.ObjAsRuntime().Spec.Shoot.SecretBindingName))

	return nil, ctx
}
