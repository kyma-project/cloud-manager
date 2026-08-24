package redisinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

// removeReadyCondition clears the Ready condition at the start of a delete
// flow so the SKR side observes the instance leaving the Ready state.
func removeReadyCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	kcp := state.ObjAsRedisInstance()
	if meta.FindStatusCondition(kcp.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) == nil {
		return nil, ctx
	}
	meta.RemoveStatusCondition(&kcp.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if err := state.UpdateObjStatus(ctx); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error removing Ready condition from AliCloud RedisInstance",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	return nil, ctx
}
