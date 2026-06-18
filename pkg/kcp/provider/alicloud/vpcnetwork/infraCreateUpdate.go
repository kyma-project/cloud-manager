package vpcnetwork

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func infraCreateUpdate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	name := state.ObjAsVpcNetwork().Status.Identifiers.Name
	if name == "" {
		return nil, ctx
	}

	// check if VPC already exists
	vpcs, err := state.alicloudClient.DescribeVpcs(ctx, name)
	if err != nil {
		logger.Error(err, "Error describing AliCloud VPCs")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	if len(vpcs) > 0 {
		// VPC already exists - check if it's available
		if vpcs[0].Status != "Available" {
			logger.Info("AliCloud VPC not yet available, requeueing", "status", vpcs[0].Status)
			return composed.StopWithRequeue, ctx
		}
		// VPC is available, update status
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.Status.Identifiers.Vpc = vpcs[0].VpcId
			}).
			OnStatusChanged(composed.Log("AliCloud VPC already exists, status updated")).
			Run(ctx, state.Cluster().K8sClient())
	}

	// create VPC
	cidr := ""
	if len(state.ObjAsVpcNetwork().Spec.CidrBlocks) > 0 {
		cidr = state.ObjAsVpcNetwork().Spec.CidrBlocks[0]
	}

	vpcInfo, err := state.alicloudClient.CreateVpc(ctx, name, cidr)
	if err != nil {
		logger.Error(err, "Error creating AliCloud VPC")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	logger.Info("AliCloud VPC created", "vpcId", vpcInfo.VpcId)

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Vpc = vpcInfo.VpcId
			vpcNetwork.Status.CidrBlocks = []string{cidr}
		}).
		OnSuccess(
			composed.Log("AliCloud VPC created, requeueing to wait for Available"),
			composed.Requeue,
		).
		Run(ctx, state.Cluster().K8sClient())
}
