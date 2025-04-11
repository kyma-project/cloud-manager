package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func gardenerClusterLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	gc := util.NewGardenerClusterUnstructured()
	err := state.Cluster().K8sClient().Get(ctx, state.Name(), gc)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading GardenerCluster", composed.StopWithRequeue, ctx)
	}

	if err == nil {
		state.gardenerCluster = gc
		summary, err := util.ExtractGardenerClusterSummary(gc)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error extracting GardenerClusterSummary", composed.StopAndForget, ctx)
		}
		state.gardenerClusterSummary = summary
	}

	return nil, ctx
}
