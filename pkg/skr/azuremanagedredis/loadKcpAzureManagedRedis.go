package azuremanagedredis

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpAzureManagedRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsAzureManagedRedis().Status.Id == "" {
		return composed.LogErrorAndReturn(
			errors.New("missing SKR AzureManagedRedis status.id"),
			"Logical error in loadKcpAzureManagedRedis",
			composed.StopAndForget,
			ctx,
		)
	}

	kcpAMR := &cloudcontrolv1beta1.AzureManagedRedis{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsAzureManagedRedis().Status.Id,
	}, kcpAMR)
	if apierrors.IsNotFound(err) {
		state.KcpAzureManagedRedis = nil
		logger.Info("KCP AzureManagedRedis does not exist")
		return nil, ctx
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP AzureManagedRedis", composed.StopWithRequeue, ctx)
	}

	state.KcpAzureManagedRedis = kcpAMR
	return nil, ctx
}
