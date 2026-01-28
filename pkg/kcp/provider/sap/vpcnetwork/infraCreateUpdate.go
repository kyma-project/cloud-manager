package vpcnetwork

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func infraCreateUpdate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	name := fmt.Sprintf("cm-%s", state.ObjAsVpcNetwork().Name)

	out, err := CreateInfra(ctx, WithName(name), WithCidrBlocks(state.ObjAsVpcNetwork().Spec.CidrBlocks), WithClient(state.sapClient))

	if err != nil {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(err.Error())
			}).
			OnSuccess(
				composed.LogError(err, "Provider error on KCP SAP VpcNetwork create/update"),
				composed.Requeue,
			).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Vpc = out.InternalNetwork.ID
			vpcNetwork.Status.Identifiers.Router = out.Router.ID
		}).
		OnSuccess(
			// log only if something was created/updated
			composed.LogIf(out.Updated, "SAP KCP VpcNetwork is successfully updated"),
			composed.LogIf(out.Created, "SAP KCP VpcNetwork is successfully created"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
