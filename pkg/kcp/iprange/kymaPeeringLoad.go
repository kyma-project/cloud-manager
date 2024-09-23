package iprange

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func kymaPeeringLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx).
		WithValues(
			"kcpIpRangeVpcPeeringName", state.Scope().Name,
		)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	peering := &cloudcontrolv1beta1.VpcPeering{}
	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{
		Namespace: state.ObjAsIpRange().Namespace,
		Name:      state.Scope().Name,
	}, peering)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP IpRange kyma peering", composed.StopWithRequeue, ctx)
	}

	if err == nil {
		logger.Info("KCP IpRange VpcPeering loaded")
		state.kymaPeering = peering
	} else {
		logger.Info("KCP IpRange VpcPeering does not exist")
	}

	return nil, ctx
}
