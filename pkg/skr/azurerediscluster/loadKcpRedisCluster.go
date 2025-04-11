package azurerediscluster

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsAzureRedisCluster().Status.Id == "" {
		return composed.LogErrorAndReturn(
			errors.New("missing SKR AzureRedisCluster state.id"),
			"Logical error in loadKcpRedisCluster",
			composed.StopAndForget,
			ctx,
		)
	}

	kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsAzureRedisCluster().Status.Id,
	}, kcpRedisCluster)
	if apierrors.IsNotFound(err) {
		state.KcpRedisCluster = nil
		logger.Info("KCP RedisCluster does not exist")
		return nil, nil
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP RedisCluster", composed.StopWithRequeue, ctx)
	}

	state.KcpRedisCluster = kcpRedisCluster

	return nil, nil
}
