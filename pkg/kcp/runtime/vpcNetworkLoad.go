package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func vpcNetworkLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsRuntime().Spec.Shoot.Networking.VPCNetwork == nil {
		return nil, ctx
	}

	vpcNetworkName := ptr.Deref(state.ObjAsRuntime().Spec.Shoot.Networking.VPCNetwork, "")
	if vpcNetworkName == "" {
		return nil, ctx
	}

	vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      vpcNetworkName,
	}, vpcNetwork)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Failed to load Runtime's VpcNetwork", composed.StopWithRequeueDelay(rate.Slow1s.When(state.ObjAsRuntime())), ctx)
	}
	if err == nil {
		state.vpcNetwork = vpcNetwork
	}

	return nil, ctx
}
