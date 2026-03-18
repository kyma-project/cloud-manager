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

	out, err := CreateInfra(
		ctx,
		WithName(name),
		WithGcpProjectId(state.Subscription().Status.SubscriptionInfo.Gcp.Project),
		WithClient(state.gcpClient),
		WithRegion(state.ObjAsVpcNetwork().Spec.Region),
	)

	if err != nil {
		return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
			MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
				vpcNetwork.SetStatusProviderError(err.Error())
			}).
			OnSuccess(
				composed.LogError(err, "Provider error on KCP VpcNetwork create/update"),
				composed.Requeue,
			).
			Run(ctx, state.Cluster().K8sClient())
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Vpc = out.Network.GetName()
			vpcNetwork.Status.Identifiers.Router = out.Router.GetName()
		}).
		OnSuccess(
			// log only if something was created
			composed.LogIf(out.Created, "GCP KCP VpcNetwork is successfully created"),
		).
		Run(ctx, state.Cluster().K8sClient())
}
