package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shareUpdateStatusCapacity(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsNfsInstance().Status.CapacityGb == state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb {
		return nil, nil
	}

	state.ObjAsNfsInstance().Status.CapacityGb = state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		SuccessErrorNil().
		ErrorLogMessage("Error updating CCEE NfsInstance status capacity").
		Run(ctx, state)
}
