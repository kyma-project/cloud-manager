package iprange

import (
	"context"
	"fmt"
	iprangetypes "github.com/kyma-project/cloud-resources/components/kcp/pkg/iprange/types"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		state, err := stateFactory.NewState(ctx, st.(iprangetypes.State))
		if err != nil {
			err = fmt.Errorf("error creating new aws iprange state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}
		return composed.ComposeActions(
			"awsIpRange",
			splitRangeByZones,
			ensureShootZonesAndRangeSubnetsMatch,
			loadVpc,
			checkCidrOverlap,
			checkCidrBlockStatus,
			extendVpcAddressSpace,
			loadSubnets,
			findCloudResourceSubnets,
			createSubnets,
			updateSuccessStatus,
			func(_ context.Context, _ composed.State) (error, context.Context) {
				return composed.StopAndForget, nil
			},
		)(ctx, state)
	}
}
