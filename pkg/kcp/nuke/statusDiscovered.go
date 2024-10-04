package nuke

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"time"
)

func statusDiscovered(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	if state.ObjAsNuke().Status.State == "" {
		changed = true
		state.ObjAsNuke().Status.State = "Processing"
	}

	if state.ObjAsNuke().Status.InitializedOn.IsZero() {
		state.ObjAsNuke().Status.InitializedOn = ptr.To(metav1.NewTime(time.Now()))
		changed = true
	}

	for _, rks := range state.Resources {
		kindStatus, created := state.ObjAsNuke().Status.GetKind(rks.Kind)
		if created {
			changed = true
		}

		for _, obj := range rks.Objects {
			_, exists := kindStatus.Objects[obj.GetName()]
			if !exists {
				changed = true
				kindStatus.Objects[obj.GetName()] = cloudcontrolv1beta1.NukeResourceStatusDiscovered
			}
		}
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsNuke()).
		ErrorLogMessage("Error patching KCP Nuke status with discovered resources").
		SuccessErrorNil().
		Run(ctx, state)
}
