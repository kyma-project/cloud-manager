package redisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisInstance := state.ObjAsRedisInstance()

	if state.gcpRedisInstance != nil {
		return nil, nil
	}

	logger.Info("Creating GCP Redis")

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	vpcNetworkFullName := fmt.Sprintf("projects/%s/global/networks/%s", gcpScope.Project, gcpScope.VpcNetwork)

	redisInstanceOptions := client.CreateRedisInstanceOptions{
		VPCNetworkFullName:    vpcNetworkFullName,
		IPRangeName:           state.IpRange().Spec.RemoteRef.Name,
		MemorySizeGb:          redisInstance.Spec.Instance.Gcp.MemorySizeGb,
		Tier:                  redisInstance.Spec.Instance.Gcp.Tier,
		RedisVersion:          redisInstance.Spec.Instance.Gcp.RedisVersion,
		AuthEnabled:           redisInstance.Spec.Instance.Gcp.AuthEnabled,
		TransitEncryptionMode: redisInstance.Spec.Instance.Gcp.TransitEncryptionMode,
		RedisConfigs:          redisInstance.Spec.Instance.Gcp.RedisConfigs.ToMap(),
	}

	_, err := state.memorystoreClient.CreateRedisInstance(ctx, gcpScope.Project, region, state.GetRemoteRedisName(), redisInstanceOptions)

	if err != nil {
		logger.Error(err, "Error creating GCP Redis")
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed creating GcpRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed gcp redis creation",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
