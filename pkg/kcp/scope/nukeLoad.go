package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func nukeLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !cloudcontrolv1beta1.AutomaticNuke {
		return nil, ctx
	}

	nuke := &cloudcontrolv1beta1.Nuke{}
	err := state.Cluster().K8sClient().Get(ctx, state.Name(), nuke)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading Nuke", composed.StopWithRequeue, ctx)
	}
	if err == nil {
		state.nuke = nuke
	}

	return nil, ctx
}
