package alicloudrediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

// waitKcpStatusUpdate waits until the KCP RedisCluster has received at least
// one terminal condition (Ready or Error) before continuing. Checking condition
// presence is more reliable than checking conditions length, since the object
// could have conditions from a previous lifecycle that don't reflect the current
// create operation.
func waitKcpStatusUpdate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.KcpRedisCluster == nil {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}
	conditions := state.KcpRedisCluster.Status.Conditions

	hasReady := meta.FindStatusCondition(conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasError := meta.FindStatusCondition(conditions, cloudcontrolv1beta1.ConditionTypeError) != nil

	if hasReady || hasError {
		return nil, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
