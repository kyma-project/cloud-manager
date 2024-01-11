package iprange

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/components/kcp/pkg/iprange/types"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		logger := composed.LoggerFromCtx(ctx)
		state, err := stateFactory.NewState(ctx, st.(types.State))
		if err != nil {
			err = fmt.Errorf("error creating new gcp iprange state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}
		return composed.ComposeActions(
			"gcpIpRange",
			loadAddress,
			loadPsaConnection,
			checkNupdateStatus,
			syncAddress,
			syncPsaConnection,
		)(ctx, state)
	}
}
