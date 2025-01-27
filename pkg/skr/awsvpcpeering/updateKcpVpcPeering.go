package awsvpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateKcpVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	obj := state.ObjAsAwsVpcPeering()

	if composed.IsMarkedForDeletion(obj) {
		return nil, nil
	}

	if state.KcpVpcPeering == nil {
		return nil, nil
	}

	shouldUpdate := false

	if state.KcpVpcPeering.Spec.Details.RemoteRouteTableUpdateStrategy != cloudcontrolv1beta1.AwsRouteTableUpdateStrategy(obj.Spec.RemoteRouteTableUpdateStrategy) {
		state.KcpVpcPeering.Spec.Details.RemoteRouteTableUpdateStrategy = cloudcontrolv1beta1.AwsRouteTableUpdateStrategy(obj.Spec.RemoteRouteTableUpdateStrategy)
		shouldUpdate = true
	}

	if state.KcpVpcPeering.Spec.Details.DeleteRemotePeering != obj.Spec.DeleteRemotePeering {
		state.KcpVpcPeering.Spec.Details.DeleteRemotePeering = obj.Spec.DeleteRemotePeering
		shouldUpdate = true
	}

	if !shouldUpdate {
		return nil, nil
	}

	err := state.KcpCluster.K8sClient().Update(ctx, state.KcpVpcPeering)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP VpcPeering", composed.StopWithRequeue, ctx)
	}

	logger.Info("Updated KCP VpcPeering")

	return nil, nil
}
