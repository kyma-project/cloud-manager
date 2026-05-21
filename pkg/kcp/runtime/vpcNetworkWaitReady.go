package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func vpcNetworkWaitReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	readyCond := meta.FindStatusCondition(state.vpcNetwork.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond != nil && readyCond.Status == metav1.ConditionTrue {
		return nil, ctx
	}

	return composed.StopWithRequeueDelay(rate.Quick.When(state.vpcNetwork)), ctx
}
