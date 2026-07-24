package nuke

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shortCircuitCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsNuke().Status.State == "Completed" {
		return composed.StopAndForget, nil
	}

	return nil, ctx
}
