package gcpvpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpGcpVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	kcpVpcPeering := &cloudcontrolv1beta1.VpcPeering{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsGcpVpcPeering().Status.Id,
	}, kcpVpcPeering)

	if apierrors.IsNotFound(err) {
		state.KcpVpcPeering = nil
		logger.Info("KCP GcpVpcPeering does not exist ", "kcpGcpVpcPeering", "cm-"+state.ObjAsGcpVpcPeering().Status.Id)
		return nil, nil
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP GcpVpcPeering "+state.ObjAsGcpVpcPeering().Status.Id, composed.StopWithRequeue, ctx)
	}

	state.KcpVpcPeering = kcpVpcPeering

	return nil, nil
}
