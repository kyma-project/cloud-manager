package cceenfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func stopIfNotBeingDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	if !composed.IsMarkedForDeletion(st.Obj()) {
		return composed.StopAndForget, ctx
	}

	return nil, ctx
}
