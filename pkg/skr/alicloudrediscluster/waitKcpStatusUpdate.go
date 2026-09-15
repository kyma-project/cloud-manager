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
// When Ready, it also waits for DiscoveryEndpoint to be non-empty: the KCP
// reconciler sets the endpoint and the Ready condition in the same cycle, so
// the cache may still show an empty endpoint on the very next SKR reconcile.
// Requeueing here ensures the auth secret is always created with a valid host.
func waitKcpStatusUpdate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.KcpRedisCluster == nil {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}
	conditions := state.KcpRedisCluster.Status.Conditions

	hasReady := meta.FindStatusCondition(conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasError := meta.FindStatusCondition(conditions, cloudcontrolv1beta1.ConditionTypeError) != nil

	if hasError {
		return nil, ctx
	}

	if hasReady {
		if state.KcpRedisCluster.Status.DiscoveryEndpoint == "" {
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}
		return nil, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
