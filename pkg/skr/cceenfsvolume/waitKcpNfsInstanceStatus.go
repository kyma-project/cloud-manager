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

	if meta.FindStatusCondition(*state.ObjAsCceeNfsVolume().Conditions(), cloudcontrol1beta1.ConditionTypeReady) != nil {
		return nil, ctx
	}
	if meta.FindStatusCondition(*state.ObjAsCceeNfsVolume().Conditions(), cloudcontrol1beta1.ConditionTypeError) != nil {
		return composed.StopAndForget, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
