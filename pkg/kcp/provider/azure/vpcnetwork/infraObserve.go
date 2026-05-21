package vpcnetwork

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func infraObserve(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	resourceGroup, err := state.azureClient.GetResourceGroup(ctx, state.ObjAsVpcNetwork().Status.Identifiers.Name)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error observing resource group")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(fmt.Sprintf("Error observing resource group: %s", err.Error()))
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	virtualNetwork, err := state.azureClient.GetNetwork(ctx, state.ObjAsVpcNetwork().Status.Identifiers.Name, state.ObjAsVpcNetwork().Status.Identifiers.Name)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error observing virtual network")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(fmt.Sprintf("Error observing virtual network: %s", err.Error()))
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Vpc = ptr.Deref(virtualNetwork.ID, "")
			vpcNetwork.Status.Identifiers.ResourceGroup = ptr.Deref(resourceGroup.ID, "")
		}).
		OnStatusChanged(
			composed.Log("Observed Azure VPC Network type Gardener status patched"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
