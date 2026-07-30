package rediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func addUpdatingCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisCluster()
	isChanging := state.instance.InstanceStatus == alicloudclient.InstanceStatusChanging
	hasUpdating := meta.FindStatusCondition(kcp.Status.Conditions, cloudcontrolv1beta1.ConditionTypeUpdating) != nil

	if isChanging == hasUpdating {
		return nil, ctx
	}

	if isChanging {
		return composed.UpdateStatus(kcp).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeUpdating,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeUpdating,
				Message: "AliCloud r-kvstore cluster is updating.",
			}).
			SuccessErrorNil().
			ErrorLogMessage("Error adding Updating condition to AliCloud RedisCluster").
			Run(ctx, st)
	}

	return composed.UpdateStatus(kcp).
		RemoveConditions(cloudcontrolv1beta1.ConditionTypeUpdating).
		SuccessErrorNil().
		ErrorLogMessage("Error removing Updating condition from AliCloud RedisCluster").
		Run(ctx, st)
}
