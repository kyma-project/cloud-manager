package cceenfsvolume

import (
	"context"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func waitKcpNfsInstanceStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.KcpNfsInstance == nil {
		return nil, ctx
	}

	// if stuck in creating, we want to be able to be deleted
	if composed.IsMarkedForDeletion(state.ObjAsCceeNfsVolume()) {
		return nil, ctx
	}
	if composed.IsMarkedForDeletion(state.KcpNfsInstance) {
		return nil, ctx
	}

	if meta.FindStatusCondition(*state.KcpNfsInstance.Conditions(), cloudcontrol1beta1.ConditionTypeReady) != nil {
		return nil, ctx
	}
	if meta.FindStatusCondition(*state.KcpNfsInstance.Conditions(), cloudcontrol1beta1.ConditionTypeError) != nil {
		return nil, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
}
