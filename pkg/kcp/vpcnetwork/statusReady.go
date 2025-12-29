package vpcnetwork

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.SetStatusProvisioned()
		}).
		OnSuccess(
			composed.Log("VpcNetwork is Ready"),
		).
		Run(ctx, state.Cluster().K8sClient())
}

