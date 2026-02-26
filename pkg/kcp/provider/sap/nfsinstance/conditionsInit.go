package nfsinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func conditionsInit(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false
	if state.ObjAsNfsInstance().Status.State == "" {
		changed = true
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateProcessing
	}

	if changed {
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			ErrorLogMessage("Error updating initial conditions").
			FailedError(composed.StopWithRequeue).
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, ctx
}
