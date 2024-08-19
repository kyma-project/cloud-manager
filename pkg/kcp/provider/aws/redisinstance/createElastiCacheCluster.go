package redisinstance

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createElastiCacheCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisInstance := state.ObjAsRedisInstance()

	if state.elastiCacheCluster != nil {
		return nil, nil
	}

	logger.Info("Creating AWS ElastiCache")

	var authTokenSecetString *string = nil

	if state.authTokenValue != nil {
		authTokenSecetString = state.authTokenValue.SecretString
	}

	_, err := state.awsClient.CreateElastiCacheCluster(ctx, []types.Tag{
		{
			Key:   ptr.To(common.TagCloudManagerName),
			Value: ptr.To(state.Name().String()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerRemoteName),
			Value: ptr.To(redisInstance.Spec.RemoteRef.String()),
		},
		{
			Key:   ptr.To(common.TagScope),
			Value: ptr.To(redisInstance.Spec.Scope.Name),
		},
		{
			Key:   ptr.To(common.TagShoot),
			Value: ptr.To(state.Scope().Spec.ShootName),
		},
	}, client.CreateElastiCacheClusterOptions{
		Name:                       GetAwsElastiCacheClusterName(state.Obj().GetName()),
		SubnetGroupName:            ptr.Deref(state.subnetGroup.CacheSubnetGroupName, ""),
		ParameterGroupName:         ptr.Deref(state.parameterGroup.CacheParameterGroupName, ""),
		CacheNodeType:              redisInstance.Spec.Instance.Aws.CacheNodeType,
		EngineVersion:              redisInstance.Spec.Instance.Aws.EngineVersion,
		AutoMinorVersionUpgrade:    redisInstance.Spec.Instance.Aws.AutoMinorVersionUpgrade,
		AuthTokenSecretString:      authTokenSecetString,
		TransitEncryptionEnabled:   redisInstance.Spec.Instance.Aws.TransitEncryptionEnabled,
		PreferredMaintenanceWindow: redisInstance.Spec.Instance.Aws.PreferredMaintenanceWindow,
	})

	if err != nil {
		logger.Error(err, "Error creating AWS ElastiCache")
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed creating AWS Elasticache: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed aws elasticache creation",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
