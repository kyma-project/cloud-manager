package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func stopIfNotBeingDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return composed.StopAndForget, nil
	}

	return nil, nil
}
