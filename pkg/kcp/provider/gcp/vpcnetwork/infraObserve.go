package vpcnetwork

import (
	"context"
	"fmt"

	"cloud.google.com/go/compute/apiv1/computepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
)

func infraObserve(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	projectId := state.Subscription().Status.SubscriptionInfo.Gcp.Project
	name := state.ObjAsVpcNetwork().Status.Identifiers.Name

	net, err := state.gcpClient.GetNetwork(ctx, &computepb.GetNetworkRequest{
		Project: projectId,
		Network: name,
	})
	if gcpmeta.IsNotFound(err) {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError("GCP network not found")
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error observing GCP network")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(fmt.Sprintf("Error observing GCP network: %s", err.Error()))
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	router, err := state.gcpClient.GetRouter(ctx, &computepb.GetRouterRequest{
		Project: projectId,
		Region:  state.ObjAsVpcNetwork().Spec.Region,
		Router:  RouterName(name),
	})
	if gcpmeta.IsNotFound(err) {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError("GCP router not found")
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error observing GCP router")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(fmt.Sprintf("Error observing GCP router: %s", err.Error()))
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Vpc = net.GetName()
			vpcNetwork.Status.Identifiers.Router = router.GetName()
		}).
		OnStatusChanged(
			composed.Log("Observed GCP VPC Network type Gardener status patched"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
