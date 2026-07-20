package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// loadAccessGroup loads the NAS permission group for this instance into state, if it exists.
func loadAccessGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	state.accessGroupName = state.AccessGroupName()

	groups, err := state.client.DescribeAccessGroups(ctx, state.accessGroupName)
	if err != nil {
		logger.Error(err, "Error loading AliCloud NAS access group")
		return composed.StopWithRequeue, ctx
	}

	for _, g := range groups {
		if g.AccessGroupName == state.accessGroupName {
			gCopy := g
			state.accessGroup = &gCopy
			break
		}
	}

	return nil, ctx
}
