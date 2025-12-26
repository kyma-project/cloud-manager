package commonAction

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func ipRangeLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*stateImpl)

	if ipRange, ok := state.Obj().(*cloudcontrolv1beta1.IpRange); ok {
		state.ipRange = ipRange
		return nil, ctx
	}

	var dependencyName string

	switch x := state.ObjAsObjWithStatus().(type) {
	case *cloudcontrolv1beta1.NfsInstance:
		dependencyName = x.Spec.IpRange.Name
	case *cloudcontrolv1beta1.RedisInstance:
		dependencyName = x.Spec.IpRange.Name
	case *cloudcontrolv1beta1.RedisCluster:
		dependencyName = x.Spec.IpRange.Name
	// add here any new KCP kind referring to IpRange
	default:
		return nil, ctx
	}

	ipRange := &cloudcontrolv1beta1.IpRange{}

	err, ctx := genericDependencyLoad(ctx, ipRange, state.ObjAsObjWithStatus(), state.Cluster().K8sClient(), state.Obj().GetNamespace(), dependencyName, "IpRange")

	if err == nil {
		state.ipRange = ipRange
	}

	return err, ctx
}
