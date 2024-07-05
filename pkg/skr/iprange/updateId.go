package iprange

import (
	"context"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsIpRange().Status.Id != "" {
		logger.Info("SKR IpRange already has ID set")
		return nil, nil
	}

	id := uuid.NewString()

	if state.ObjAsIpRange().Labels == nil {
		state.ObjAsIpRange().Labels = map[string]string{}
	}

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR IpRange with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR IpRange updated with ID label")

	state.ObjAsIpRange().Status.Id = id
	state.ObjAsIpRange().SetState(cloudresourcesv1beta1.StateProcessing)
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR IpRange status with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR IpRange updated with ID status")

	return composed.StopWithRequeue, nil
}
