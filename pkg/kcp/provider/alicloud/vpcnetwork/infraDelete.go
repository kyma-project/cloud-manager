package vpcnetwork

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func infraDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	vpcId := state.ObjAsVpcNetwork().Status.Identifiers.Vpc
	if vpcId == "" {
		// no VPC to delete
		return nil, ctx
	}

	name := state.ObjAsVpcNetwork().Status.Identifiers.Name

	// verify VPC still exists
	vpcs, err := state.alicloudClient.DescribeVpcs(ctx, name)
	if err != nil {
		logger.Error(err, "Error describing AliCloud VPCs during delete")
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	if len(vpcs) == 0 {
		logger.Info("AliCloud VPC already deleted", "vpcId", vpcId)
		return nil, ctx
	}

	err = state.alicloudClient.DeleteVpc(ctx, vpcId)
	if err != nil {
		logger.Error(err, "Error deleting AliCloud VPC", "vpcId", vpcId)
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	logger.Info("AliCloud VPC deleted", "vpcId", vpcId)

	return nil, ctx
}
