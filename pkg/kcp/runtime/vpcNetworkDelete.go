package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func vpcNetworkDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.vpcNetwork == nil {
		return nil, ctx
	}
	if state.vpcNetwork.Spec.Type != cloudcontrolv1beta1.VpcNetworkTypeGardener {
		return nil, ctx
	}
	if composed.IsMarkedForDeletion(state.vpcNetwork) {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info("Deleting Gardener type VpcNetwork", "vpcNetworkName", state.vpcNetwork.Name)

	err := state.Cluster().K8sClient().Delete(ctx, state.vpcNetwork)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Failed to delete Gardener type VpcNetwork", composed.StopWithRequeueDelay(rate.Quick.When(state.vpcNetwork)), ctx)
	}
	return nil, ctx
}
