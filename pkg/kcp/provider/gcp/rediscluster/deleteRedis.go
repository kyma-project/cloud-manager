package rediscluster

import (
	"context"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.gcpRedisCluster == nil {
		return nil, nil
	}

	if state.gcpRedisCluster.State == clusterpb.Cluster_DELETING {
		return nil, nil // delete is waited in next action
	}

	logger.Info("Deleting GCP Redis")

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	err := state.memorystoreClient.DeleteRedisCluster(ctx, gcpScope.Project, region, state.GetRemoteRedisName())
	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target redis instance for delete not found, continuing to next loop")
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
		}

		logger.Error(err, "Error deleting GCP Redis")
		redisCluster := state.ObjAsRedisCluster()
		meta.SetStatusCondition(redisCluster.Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonCloudProviderError,
			Message: "Failed to delete RedisCluster",
		})
		redisCluster.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed gcp redis deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
