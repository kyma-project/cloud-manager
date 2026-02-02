package vpcnetwork

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func infraCreateUpdate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	name := fmt.Sprintf("cm-%s", state.ObjAsVpcNetwork().Name)

	out, err := CreateInfra(ctx, WithName(name), WithCidrBlocks(state.ObjAsVpcNetwork().Spec.CidrBlocks), WithClient(state.awsClient))
	if err != nil {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(err.Error())
			}).
			OnSuccess(
				composed.LogError(err, "Provider error on KCP AWS VpcNetwork create/update"),
				composed.Requeue,
			).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Vpc = ptr.Deref(out.Vpc.VpcId, "")
			vpcNetwork.Status.Identifiers.InternetGateway = ptr.Deref(out.InternetGateway.InternetGatewayId, "")
		}).
		OnSuccess(
			// log only if something was created/updated
			composed.LogIf(out.Updated, "AWS KCP VpcNetwork is successfully updated"),
			composed.LogIf(out.Created, "AWS KCP VpcNetwork is successfully created"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
