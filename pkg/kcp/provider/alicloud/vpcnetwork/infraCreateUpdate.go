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
		// VPC already exists, update status
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
		OnSuccess(composed.Log("AliCloud VPC created successfully")).
		Run(ctx, state.Cluster().K8sClient())
}
