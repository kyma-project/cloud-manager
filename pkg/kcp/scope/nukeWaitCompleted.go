package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func nukeWaitCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !cloudcontrolv1beta1.AutomaticNuke {
		return nil, ctx
	}

	if state.nuke == nil {
		return nil, ctx
	}

	readyCond := meta.FindStatusCondition(state.nuke.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond != nil {
		return nil, ctx
	}

	return composed.StopWithRequeue, ctx
}
