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
	if name == "" {
		return nil, ctx
	}

	// Gardener names the VPC as "<shootName>-vpc" on Alicloud
	vpcs, err := state.alicloudClient.DescribeVpcs(ctx, name+"-vpc")
	if err != nil {
		composed.LoggerFromCtx(ctx).Error(err, "Error observing AliCloud VPC")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(fmt.Sprintf("Error observing VPC: %s", err.Error()))
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	if len(vpcs) == 0 {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError("VPC not found")
			}).
			OnSuccess(composed.RequeueAfter(rate.Slow1s.When(state.ObjAsVpcNetwork()))).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Vpc = vpcs[0].VpcId
		}).
		OnStatusChanged(
			composed.Log("Observed AliCloud VPC Network status patched"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
