package redisinstance

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// waitRedisAvailable polls the instance status until it reaches Normal. While
// Creating, Changing, or SSLModifying (transient post-create SSL configuration),
// the reconciler requeues. Any other terminal state surfaces an Error condition.
func waitRedisAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	switch state.instance.InstanceStatus {
	case alicloudclient.InstanceStatusNormal:
		return nil, ctx
	case alicloudclient.InstanceStatusCreating, alicloudclient.InstanceStatusChanging,
		alicloudclient.InstanceStatusSSLModifying:
		state.instance = nil
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
	default:
		kcp := state.ObjAsRedisInstance()
		return composed.UpdateStatus(kcp).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingRedisInstance,
				Message: fmt.Sprintf("AliCloud r-kvstore instance in unexpected state: %s", state.instance.InstanceStatus),
			}).
			ErrorLogMessage("Error updating KCP RedisInstance status for unexpected AliCloud state").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}
}
