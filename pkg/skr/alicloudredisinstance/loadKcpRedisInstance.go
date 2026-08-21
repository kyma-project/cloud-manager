package alicloudredisinstance

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsAlicloudRedisInstance().Status.Id == "" {
		if composed.MarkedForDeletionPredicate(ctx, st) {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(
			errors.New("SKR AlicloudRedisInstance has no status.id"),
			"SKR AlicloudRedisInstance has no status.id - requeuing",
			composed.StopWithRequeue,
			ctx,
		)
	}

	kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsAlicloudRedisInstance().Status.Id,
	}, kcpRedisInstance)
	if apierrors.IsNotFound(err) {
		state.KcpRedisInstance = nil
		logger.Info("KCP RedisInstance does not exist")
		return nil, ctx
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	state.KcpRedisInstance = kcpRedisInstance

	return nil, ctx
}
