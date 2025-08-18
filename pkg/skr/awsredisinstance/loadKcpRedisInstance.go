package awsredisinstance

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

	if state.ObjAsAwsRedisInstance().Status.Id == "" {
		return composed.LogErrorAndReturn(
			errors.New("missing SKR AwsRedisInstance state.id"),
			"Logical error in loadKcpRedisInstance",
			composed.StopAndForget,
			ctx,
		)
	}

	kcpRedisInstnace := &cloudcontrolv1beta1.RedisInstance{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsAwsRedisInstance().Status.Id,
	}, kcpRedisInstnace)
	if apierrors.IsNotFound(err) {
		state.KcpRedisInstance = nil
		logger.Info("KCP RedisInstance does not exist")
		return nil, ctx
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	state.KcpRedisInstance = kcpRedisInstnace

	return nil, ctx
}
