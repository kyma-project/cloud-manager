package azurevpcdnslink

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpAzureVNetLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsVNetLink().Status.Id == "" {
		return composed.LogErrorAndReturn(
			common.ErrLogical,
			"Missing SKR AzureVNetLink state.id",
			composed.StopAndForget,
			ctx,
		)
	}

	kcpVpcPeering := &cloudcontrolv1beta1.AzureVNetLink{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsVNetLink().Status.Id,
	}, kcpVpcPeering)

	if apierrors.IsNotFound(err) {
		state.KcpAzureVNetLink = nil
		logger.Info("KCP AzureVNetLink does not exist")
		return nil, ctx
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP AzureVNetLink", composed.StopWithRequeue, ctx)
	}

	state.KcpAzureVNetLink = kcpVpcPeering

	return nil, ctx
}
