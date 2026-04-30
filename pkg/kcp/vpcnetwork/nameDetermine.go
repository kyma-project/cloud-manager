package vpcnetwork

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	vpcnetworkconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/config"
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
		if vpcnetworkconfig.VpcNetworkConfig.Prefix != "" {
			name = fmt.Sprintf("kyma-%s-%s", vpcnetworkconfig.VpcNetworkConfig.Prefix, state.ObjAsVpcNetwork().Name)
		} else {
			name = fmt.Sprintf("kyma-%s", state.ObjAsVpcNetwork().Name)
		}
		if len(name) > 60 {
			name = name[:60]
		}
	}

	return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
		MutateStatus(func(vpcNetwork *cloudcontrolv1beta1.VpcNetwork) {
			vpcNetwork.Status.Identifiers.Name = name
		}).
		Run(ctx, state.Cluster().K8sClient())
}
