package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func loadRemoteRouteTables(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// remote client not created
	if state.remoteClient == nil {
		return nil, nil
	}

	if state.remoteVpc == nil {
		return nil, nil
	}

	routeTables, err := state.remoteClient.DescribeRouteTables(ctx, *state.remoteVpc.VpcId)

	if err != nil {
		if composed.IsMarkedForDeletion(state.Obj()) {
			return composed.LogErrorAndReturn(err,
				"Error loading AWS remote route tables but skipping as marked for deletion",
				nil,
				ctx)
		}
		return awsmeta.LogErrorAndReturn(err, "Error loading AWS remote route tables", ctx)
	}

	state.remoteRouteTables = routeTables

	return nil, nil
}
