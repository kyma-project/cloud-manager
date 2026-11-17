package cloudresources

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	cloudresourcestypes "github.com/kyma-project/cloud-manager/pkg/skr/cloudresources/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		crState := st.(cloudresourcestypes.State)
		state, err := stateFactory.NewState(ctx, crState)
		if err != nil {
			logger.Error(err, "error creating new aws cloudresources action")
			return composed.StopWithRequeueDelay(rate.Default.UltraSlow1m.When(st.Obj())), ctx
		}

		return composed.ComposeActionsNoName(
			logGroupLoad,
			composed.If(
				composed.MarkedForDeletionPredicate,
				logGroupDelete,
			),
			composed.If(
				composed.NotMarkedForDeletionPredicate,
				logGroupCreate,
				logGroupPutRetention,
			),
		)(ctx, state)
	}
}
