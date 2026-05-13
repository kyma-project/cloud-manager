package vpcnetwork

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func nameDetermine(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsVpcNetwork().Status.Identifiers.Name != "" {
		return nil, ctx
	}

	// name is not determined, find it now, set it to status

	name := ptr.Deref(state.ObjAsVpcNetwork().Spec.VpcNetworkName, "")
	if name == "" {
		// name is not specified
		name = common.KymaVpcName(state.ObjAsVpcNetwork().Name)
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Name = name
		}).
		Run(ctx, state.Cluster().K8sClient())
}
