package commonAction

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func gcpSubnetLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*stateImpl)

	if gcpSubnet, ok := state.Obj().(*cloudcontrolv1beta1.GcpSubnet); ok {
		state.gcpSubnet = gcpSubnet
		return nil, ctx
	}

	var dependencyName string

	switch x := state.ObjAsObjWithStatus().(type) {
	case *cloudcontrolv1beta1.GcpRedisCluster:
		dependencyName = x.Spec.Subnet.Name
	// add here any new KCP kind referring to GcpSubnet
	default:
		return nil, ctx
	}

	gcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}

	err, ctx := genericDependencyLoad(ctx, gcpSubnet, state.ObjAsObjWithStatus(), state.Cluster().K8sClient(), state.Obj().GetNamespace(), dependencyName, "GcpSubnet")

	if err == nil {
		state.gcpSubnet = gcpSubnet
	}

	return err, ctx
}
