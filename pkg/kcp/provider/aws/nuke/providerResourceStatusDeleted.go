package nuke

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func providerResourceStatusDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	for _, sk := range state.ObjAsNuke().Status.Resources {
		if sk.GetResourceType() == cloudcontrolv1beta1.ProviderResource {
			for objId, objStatus := range sk.Objects {
				if objStatus == cloudcontrolv1beta1.NukeResourceStatusDeleted {
					continue
				}
				if !state.ProviderObjectExists(sk.Kind, objId) {
					changed = true
					sk.Objects[objId] = cloudcontrolv1beta1.NukeResourceStatusDeleted
				}
			}
		}
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsNuke()).
		ErrorLogMessage("Error patching KCP Nuke status with deleted provider resources").
		SuccessErrorNil().
		Run(ctx, state)
}

func (s *State) ProviderObjectExists(kind, id string) bool {
	for _, res := range s.ProviderResources {
		if res.Kind == kind {
			for _, obj := range res.Objects {
				if obj.GetId() == id {
					return true
				}
			}
		}
	}
	return false
}
