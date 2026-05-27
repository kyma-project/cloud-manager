package commonAction

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func subscriptionLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*stateImpl)

	if subscription, ok := state.Obj().(*cloudcontrolv1beta1.Subscription); ok {
		state.subscription = subscription
		return nil, ctx
	}

	dependencyName := state.vpcNetwork.Spec.Subscription

	subscription := &cloudcontrolv1beta1.Subscription{}

	err, ctx := genericDependencyLoad(ctx, subscription, state.ObjAsObjWithStatus(), state.Cluster().K8sClient(), state.Obj().GetNamespace(), dependencyName, "Subscription")

	if err == nil {
		state.subscription = subscription
	}

	return err, ctx
}
