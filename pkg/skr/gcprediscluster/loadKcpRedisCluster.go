package gcprediscluster

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpGcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsGcpRedisCluster().Status.Id == "" {
		return composed.LogErrorAndReturn(
			errors.New("missing SKR GcpRedisCluster state.id"),
			"Logical error in loadKcpGcpRedisCluster",
			composed.StopAndForget,
			ctx,
		)
	}

	kcpGcpRedisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsGcpRedisCluster().Status.Id,
	}, kcpGcpRedisCluster)
	if apierrors.IsNotFound(err) {
		state.KcpGcpRedisCluster = nil
		logger.Info("KCP GcpRedisCluster does not exist")
		return nil, nil
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP GcpRedisCluster", composed.StopWithRequeue, ctx)
	}

	state.KcpGcpRedisCluster = kcpGcpRedisCluster

	return nil, nil
}
