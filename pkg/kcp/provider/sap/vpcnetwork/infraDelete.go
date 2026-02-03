package vpcnetwork

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func infraDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	name := fmt.Sprintf("cm-%s", state.ObjAsVpcNetwork().Name)

	err := DeleteInfra(ctx, name, state.sapClient)
	if err != nil {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(err.Error())
			}).
			OnSuccess(
				composed.LogError(err, "Provider error on KCP VpcNetwork delete"),
				composed.Requeue,
			).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers = cloudcontrolv1beta1.VpcNetworkStatusIdentifiers{}
		}).
		Run(ctx, state.Cluster().K8sClient())
}
