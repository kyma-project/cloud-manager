package cloudresources

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
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
			composed.If(
				composed.MarkedForDeletionPredicate,
				composed.If(
					isWafFeatureEnabled,
					logGroupLoad,
					logGroupDelete,
				),
			),
			composed.If(
				composed.NotMarkedForDeletionPredicate,
				composed.If(
					composed.All(
						isWafFeatureEnabled,
						shouldWafLogGroupBeCreatedOrUpdated,
					),
					logGroupLoad,
					logGroupCreate,
					logGroupPutRetention,
					logGroupStatus,
				),
			),
		)(ctx, state)
	}
}

// wafLogGroupImplementation is identifier of the WAF Log Group implementation
// Once done this value is saved to cloudresources.status.subscription.wafLogGroup
// Since SKR manager is periodically reconnected and does full sync often, this is used
// as a signal that log group is already created. If implementation changes, and this controller
// needs a new pass, then change the value of this constant so when composed with the
// cloudresources.status.subscription.wafLogGroup it will yield not equal status and gate opens
const wafLogGroupImplementation = "v1"

func isWafFeatureEnabled(ctx context.Context, st composed.State) bool {
	state := st.(*State)
	_, cctx := feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.WafPolicy{})(ctx, state)
	disabled := feature.ApiDisabled.Value(cctx)
	return !disabled
}

func shouldWafLogGroupBeCreatedOrUpdated(ctx context.Context, st composed.State) bool {
	state := st.(*State)
	return state.ObjAsCloudResources().Status.Subscription.WafLogGroup != wafLogGroupImplementation
}
