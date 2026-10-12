package certificate

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

	if state.ObjAsCertificate().Status.Id != "" {
		return nil, ctx
	}

	id := uuid.NewString()

	if state.ObjAsCertificate().Labels == nil {
		state.ObjAsCertificate().Labels = map[string]string{}
	}
	state.ObjAsCertificate().Labels[cloudresourcesv1beta1.LabelId] = id

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR Certificate with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR Certificate updated with ID label")

	state.ObjAsCertificate().Status.Id = id
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR Certificate status with ID", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR Certificate updated with ID status")

	return composed.StopWithRequeueDelay(100 * time.Millisecond), ctx
}
