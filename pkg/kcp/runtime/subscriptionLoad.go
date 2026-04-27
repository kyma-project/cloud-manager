package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
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
		return err, ctx
	}
	if err != nil {
		subscription = nil
	}

	state.Subscription = subscription

	ctx = composed.LoggerIntoCtx(ctx, composed.LoggerFromCtx(ctx).WithValues("subscription", state.ObjAsRuntime().Spec.Shoot.SecretBindingName))

	return nil, ctx
}
