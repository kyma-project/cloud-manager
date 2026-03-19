package awswebacl

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	// For now, no ID management needed for WebACL
	// This can be extended when KCP resource is added
	return nil, nil
}
