package nuke

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func providerResourceStatusDiscovered(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	for _, prks := range state.ProviderResources {
		kindStatus, created := state.ObjAsNuke().Status.GetKind(prks.Kind, cloudcontrolv1beta1.ProviderResource)
		if created {
			changed = true
		}

		for _, obj := range prks.Objects {
			_, exists := kindStatus.Objects[obj.GetId()]
			if !exists {
				changed = true
				kindStatus.Objects[obj.GetId()] = cloudcontrolv1beta1.NukeResourceStatusDiscovered
			}
		}
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsNuke()).
		ErrorLogMessage("Error patching KCP Nuke status with discovered provider resources").
		SuccessErrorNil().
		Run(ctx, state)
}
