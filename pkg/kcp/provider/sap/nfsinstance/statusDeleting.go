package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusDeleting(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsNfsInstance().Status.State == "Deleting" {
		return nil, nil
	}

	state.ObjAsNfsInstance().Status.State = "Deleting"

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error patching SAP NfsInstance with deleting status state").
		SuccessErrorNil().
		Run(ctx, state)
}
