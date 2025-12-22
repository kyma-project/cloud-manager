package commonAction

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
)

func subscriptionLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*stateImpl)

	if obj, ok := state.Obj().(ObjReferringSubscription); ok {
		subscription := &cloudcontrolv1beta1.Subscription{}
		err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.SubscriptionName(),
		}, subscription)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error loading Subscription", composed.StopWithRequeue, ctx)
		}
		state.subscription = subscription
	}

	return nil, ctx
}
