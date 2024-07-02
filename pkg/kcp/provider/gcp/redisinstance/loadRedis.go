package redisinstance

import (
	"context"
	"fmt"

	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.gcpRedisInstance != nil {
		logger.Info("GCP Redis already loaded")
		return nil, nil
	}

	logger.Info("Loading GCP Redis")

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	redisInstance, redisAuth, err := state.memorystoreClient.GetRedisInstance(ctx, gcpScope.Project, region, state.GetRemoteRedisName())

	if err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			if apiErr.GRPCStatus().Code() == codes.NotFound {
				return nil, nil
			}
		}

		logger.Error(err, "Error loading GCP Redis")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed loading GcpRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed gcp redis loading",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	state.gcpRedisInstance = redisInstance
	state.gcpRedisInstanceAuth = redisAuth

	return nil, nil
}
