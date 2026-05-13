package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	commongardener "github.com/kyma-project/cloud-manager/pkg/common/gardener"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func vpcNetworkCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.vpcNetwork != nil {
		return nil, ctx
	}

	name := ptr.Deref(state.ObjAsRuntime().Spec.Shoot.Networking.VPCNetwork, "")
	if name != "" && name != state.ObjAsRuntime().Name {
		// it's specified to something different from runtime id
		// that's the kyma network case, do nothing, skip
		return nil, ctx
	}

	// runtime's VPCNetwork is either empty or equal to its id (name)
	// since it's not loaded, probably doesn't exist, so will be created now

	ns, err := commongardener.DefaultGardenerNamespaceProvider().GetGardenerNamespace(ctx, state.Cluster().ApiReader())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting gardener namespace", composed.StopWithRequeueDelay(rate.Quick.When(state.ObjAsRuntime())), ctx)
	}

	vpcNetwork := cloudcontrolv1beta1.NewVpcNetworkBuilder().
		WithNamespace(state.ObjAsRuntime().Namespace).
		WithName(state.ObjAsRuntime().Name).
		WithVpcNetworkName(new(common.GardenerVpcName(ns, state.ObjAsRuntime().Spec.Shoot.Name))).
		WithRegion(state.ObjAsRuntime().Spec.Shoot.Region).
		WithSubscription(state.subscription.Name).
		WithCidrBlocks(state.ObjAsRuntime().Spec.Shoot.Networking.Nodes).
		WithType(cloudcontrolv1beta1.VpcNetworkTypeGardener).
		Build()

	if vpcNetwork.Labels == nil {
		vpcNetwork.Labels = map[string]string{}
	}
	for _, labelName := range cloudcontrolv1beta1.ScopeLabels {
		val, ok := state.ObjAsRuntime().Labels[labelName]
		if ok {
			vpcNetwork.Labels[labelName] = val
		}
	}
	for k, v := range util.NewLabelBuilder().WithCloudManagerDefaults().Build() {
		vpcNetwork.Labels[k] = v
	}

	err = state.Cluster().K8sClient().Create(ctx, vpcNetwork)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Failed to create VpcNetwork", composed.StopWithRequeueDelay(rate.Quick.When(state.ObjAsRuntime())), ctx)
	}
	state.vpcNetwork = vpcNetwork

	if name == "" {
		// not set in runtime, patch it
		original := state.ObjAsRuntime().DeepCopy()
		state.ObjAsRuntime().Spec.Shoot.Networking.VPCNetwork = new(state.ObjAsRuntime().Name)
		err = state.Cluster().K8sClient().Patch(ctx, state.ObjAsRuntime(), client.MergeFrom(original))
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error patching Runtime with Gardener type VpcNetwork name", composed.StopWithRequeueDelay(rate.Quick.When(state.ObjAsRuntime())), ctx)
		}
		composed.LoggerFromCtx(ctx).Info("Runtime patched with Gardener type VpcNetwork name")
	}

	return nil, ctx
}
