package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func gardenerClusterExtractShootName(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.gardenerCluster == nil {
		return nil, ctx
	}

	state.shootName = state.gardenerCluster.GetLabels()[cloudcontrolv1beta1.LabelScopeShootName]
	if state.shootName == "" {
		state.shootName = state.gardenerClusterSummary.Shoot
	}

	return nil, ctx
}
