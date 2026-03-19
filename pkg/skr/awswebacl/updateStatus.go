package awswebacl

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	// Status update logic will be implemented when KCP resource is added
	return nil, nil
}
