package vpcnetwork

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func infraObserve(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	name := state.ObjAsVpcNetwork().Status.Identifiers.Name

	network, err := state.sapClient.GetNetworkByName(ctx, name)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error observing OpenStack network")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(fmt.Sprintf("Error observing OpenStack network: %s", err.Error()))
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}
	if network == nil {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError("OpenStack network not found")
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	router, err := state.sapClient.GetRouterByName(ctx, name)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error observing OpenStack router")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(fmt.Sprintf("Error observing OpenStack router: %s", err.Error()))
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}
	if router == nil {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError("OpenStack router not found")
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Vpc = network.ID
			vpcNetwork.Status.Identifiers.Router = router.ID
		}).
		OnStatusChanged(
			composed.Log("Observed SAP OpenStack VPC Network type Gardener status patched"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
