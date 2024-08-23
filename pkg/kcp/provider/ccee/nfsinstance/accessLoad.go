package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func accessLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.share == nil {
		return nil, nil
	}

	arr, err := state.cceeClient.ListShareAccessRules(ctx, state.share.ID)
	if err != nil {
		return composed.LogErrorAndReturn(err, "error listing CCEE share access rights", composed.StopWithRequeue, ctx)
	}

	accessRightId, _ := state.ObjAsNfsInstance().GetStateData(StateDataAccessRightId)
	if accessRightId == "" {
		for _, accessRight := range arr {
			if accessRight.AccessTo == state.Scope().Spec.Scope.OpenStack.Network.Nodes {
				state.accessRight = &accessRight
				break
			}
		}
	} else {
		for _, accessRight := range arr {
			if accessRight.ID == accessRightId {
				state.accessRight = &accessRight
				break
			}
		}
	}

	return nil, nil
}
