package runtime

import (
	"context"
	"maps"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func subscriptionCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.subscription != nil {
		return nil, ctx
	}

	subscription := &cloudcontrolv1beta1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: state.Obj().GetNamespace(),
			Name:      state.ObjAsRuntime().Spec.Shoot.SecretBindingName,
			Labels: map[string]string{
				cloudcontrolv1beta1.SubscriptionLabelBindingName: state.ObjAsRuntime().Spec.Shoot.SecretBindingName,
				cloudcontrolv1beta1.LabelScopeProvider:           state.ObjAsRuntime().Spec.Shoot.Provider.Type,
			},
		},
		Spec: cloudcontrolv1beta1.SubscriptionSpec{
			Details: cloudcontrolv1beta1.SubscriptionDetails{
				Garden: &cloudcontrolv1beta1.SubscriptionGarden{
					BindingName: state.ObjAsRuntime().Spec.Shoot.SecretBindingName,
				},
			},
		},
	}

	if subscription.Labels == nil {
		subscription.Labels = map[string]string{}
	}
	for _, labelName := range cloudcontrolv1beta1.ScopeLabels {
		val, ok := state.ObjAsRuntime().Labels[labelName]
		if ok {
			subscription.Labels[labelName] = val
		}
	}
	maps.Copy(subscription.Labels, util.NewLabelBuilder().WithCloudManagerDefaults().Build())

	err := state.Cluster().K8sClient().Create(ctx, subscription)
	if err != nil {
		return err, ctx
	}

	state.subscription = subscription

	composed.LoggerFromCtx(ctx).Info("KCP Subscription created")

	return composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx
}
