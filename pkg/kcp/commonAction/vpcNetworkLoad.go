package commonAction

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func vpcNetworkLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*stateImpl)

	if vpcNetwork, ok := state.Obj().(*cloudcontrolv1beta1.VpcNetwork); ok {
		state.vpcNetwork = vpcNetwork
		return nil, ctx
	}

	var dependencyName string

	switch x := state.Obj().(type) {
	case *cloudcontrolv1beta1.IpRange:
		dependencyName = x.Spec.VpcNetwork.Name
	case *cloudcontrolv1beta1.GcpSubnet:
		dependencyName = x.Spec.VpcNetwork.Name
	default:
		return nil, ctx
	}

	vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

	err, ctx := genericDependencyLoad(ctx, vpcNetwork, state.ObjAsObjWithStatus(), state.Cluster().K8sClient(), state.Obj().GetNamespace(), dependencyName, "VpcNetwork")

	if err == nil {
		state.vpcNetwork = vpcNetwork
	}

	return err, ctx
}
