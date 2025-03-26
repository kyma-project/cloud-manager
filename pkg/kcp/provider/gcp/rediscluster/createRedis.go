package rediscluster

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisCluster := state.ObjAsRedisCluster()

	if state.gcpRedisCluster != nil {
		return nil, nil
	}

	logger.Info("Creating GCP Redis")

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	vpcNetworkFullName := fmt.Sprintf("projects/%s/global/networks/%s", gcpScope.Project, gcpScope.VpcNetwork)

	redisClusterOptions := client.CreateRedisClusterOptions{
		VPCNetworkFullName: vpcNetworkFullName,
		NodeType:           redisCluster.Spec.Instance.Gcp.NodeType,
		ReplicaCount:       redisCluster.Spec.Instance.Gcp.ReplicasPerShard,
		ShardCount:         redisCluster.Spec.Instance.Gcp.ShardCount,
		RedisConfigs:       redisCluster.Spec.Instance.Gcp.RedisConfigs,
	}

	err := state.memorystoreClient.CreateRedisCluster(ctx, gcpScope.Project, region, state.GetRemoteRedisName(), redisClusterOptions)

	if err != nil {
		logger.Error(err, "Error creating GCP Redis")
		meta.SetStatusCondition(redisCluster.Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonCloudProviderError,
			Message: "Failed to create RedisCluster",
		})
		redisCluster.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed gcp redis creation",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
