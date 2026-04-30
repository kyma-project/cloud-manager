package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func subscriptionWaitReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	readyCond := meta.FindStatusCondition(state.Subscription.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond != nil && readyCond.Status == metav1.ConditionTrue {
		return nil, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
