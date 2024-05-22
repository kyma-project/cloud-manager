package v2

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		ipRangeState := st.(iprangetypes.State)
		state, err := stateFactory.NewState(ctx, ipRangeState, logger)
		if err != nil {
			err = fmt.Errorf("error creating new aws iprange state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"awsIpRangeI2-main",
		)(awsmeta.SetAwsAccountId(ctx, ipRangeState.Scope().Spec.Scope.Aws.AccountId), state)
	}
}
