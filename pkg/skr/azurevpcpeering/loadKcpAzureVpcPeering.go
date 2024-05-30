package azurevpcpeering

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpAzureVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsAzureVpcPeering().Status.Id == "" {
		return composed.LogErrorAndReturn(
			errors.New("missing SKR AzureVpcPeering state.id"),
			"Logical error in loadKcpNfsInstance",
			composed.StopAndForget,
			ctx,
		)
	}

	kcpVpcPeering := &cloudcontrolv1beta1.VpcPeering{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsAzureVpcPeering().Status.Id,
	}, kcpVpcPeering)

	if apierrors.IsNotFound(err) {
		state.KcpVpcPeering = nil
		logger.Info("KCP AzureVpcPeering does not exist")
		return nil, nil
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP AzureVpcPeering", composed.StopWithRequeue, ctx)
	}

	state.KcpVpcPeering = kcpVpcPeering

	return nil, nil
}
