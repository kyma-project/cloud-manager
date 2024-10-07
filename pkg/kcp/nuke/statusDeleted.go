package nuke

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	for _, sk := range state.ObjAsNuke().Status.Resources {
		for objName, objStatus := range sk.Objects {
			if objStatus == cloudcontrolv1beta1.NukeResourceStatusDeleted {
				continue
			}
			if !state.ObjectExists(sk.Kind, objName) {
				changed = true
				sk.Objects[objName] = cloudcontrolv1beta1.NukeResourceStatusDeleted
			}
		}
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsNuke()).
		ErrorLogMessage("Error patching KCP Nuke status with deleted resources").
		SuccessErrorNil().
		Run(ctx, state)
}
