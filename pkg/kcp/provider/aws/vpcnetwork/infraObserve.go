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

	vpcArr, err := state.awsClient.DescribeVpcs(ctx, state.ObjAsVpcNetwork().Status.Identifiers.Name)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error observing VPC")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(fmt.Sprintf("Error observing VPC: %s", err.Error()))
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}
	if len(vpcArr) == 0 {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError("VPC not found")
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	igwArr, err := state.awsClient.DescribeInternetGateways(ctx, state.ObjAsVpcNetwork().Status.Identifiers.Name)
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error observing internet gateway")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(fmt.Sprintf("Error observing internet gateway: %s", err.Error()))
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}
	if len(igwArr) == 0 {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError("Internet gateway not found")
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Vpc = ptr.Deref(vpcArr[0].VpcId, "")
			vpcNetwork.Status.Identifiers.InternetGateway = ptr.Deref(igwArr[0].InternetGatewayId, "")
		}).
		OnStatusChanged(
			composed.Log("Observed AWS VPC Network type Gardener status patched"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
