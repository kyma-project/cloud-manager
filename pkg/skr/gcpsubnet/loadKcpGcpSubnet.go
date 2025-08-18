package gcpsubnet

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpGcpSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsGcpSubnet().Status.Id == "" {
		return composed.LogErrorAndReturn(
			errors.New("missing SKR GcpSubnet state.id"),
			"Logical error in loadKcpGcpSubnet",
			composed.StopAndForget,
			ctx,
		)
	}

	kcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsGcpSubnet().Status.Id,
	}, kcpSubnet)
	if apierrors.IsNotFound(err) {
		state.KcpGcpSubnet = nil
		logger.Info("KCP GcpSubnet does not exist")
		return nil, ctx
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP GcpSubnet", composed.StopWithRequeue, ctx)
	}

	state.KcpGcpSubnet = kcpSubnet

	return nil, ctx
}
