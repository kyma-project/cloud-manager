package wafpolicy

import (
	"context"
	"time"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if state.ObjAsWafPolicy().Status.Id != "" {
		return nil, ctx
	}

	id := uuid.NewString()

	if state.ObjAsWafPolicy().Labels == nil {
		state.ObjAsWafPolicy().Labels = map[string]string{}
	}
	state.ObjAsWafPolicy().Labels[cloudresourcesv1beta1.LabelId] = id

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR WafPolicy with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR WafPolicy updated with ID label")

	state.ObjAsWafPolicy().Status.Id = id
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR WafPolicy status with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR WafPolicy updated with ID status")

	return composed.StopWithRequeueDelay(100 * time.Millisecond), ctx
}
