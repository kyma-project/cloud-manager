package network

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// logLogicalErrorOnManagedNetworkMissing is a logical safeguard that is protecting the flow. All network references
// must be reconciled before this action and stopped. If a network reference reach this action it is considered a
// logical exception and a development flow
func logLogicalErrorOnManagedNetworkMissing(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*state)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsNetwork().Spec.Network.Managed == nil {
		err := errors.New("expected managed network, but none is present in state")
		logger.Error(err, "Logical error")
		return composed.StopAndForget, nil
	}

	if state.ObjAsNetwork().Spec.Network.Reference != nil {
		err := errors.New("did not expect network reference, but it is present in state")
		logger.Error(err, "Logical error")
		return composed.StopAndForget, nil
	}

	return nil, nil
}
