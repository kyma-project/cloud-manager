package rediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.gcpRedisCluster != nil {
		logger.Info("GCP Redis already loaded")
		return nil, ctx
	}

	logger.Info("Loading GCP Redis")

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	redisCluster, err := state.memorystoreClient.GetRedisCluster(ctx, gcpScope.Project, region, state.GetRemoteRedisName())

	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target redis instance not found, continuing")
			return nil, ctx
		}

		logger.Error(err, "Error loading GCP Redis")
		redisCluster := state.ObjAsGcpRedisCluster()
		meta.SetStatusCondition(redisCluster.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to load GcpRedisCluster",
		})
		redisCluster.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating GcpRedisCluster status due failed gcp redis loading",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	if redisCluster != nil {
		logger.Info("redis instance found and loaded")
		state.gcpRedisCluster = redisCluster
	}

	return nil, ctx
}
