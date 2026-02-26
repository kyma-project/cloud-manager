package nfsinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/resource"
)

func shareUpdateStatusCapacity(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsNfsInstance().Status.CapacityGb == state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb {
		return nil, ctx
	}

	state.ObjAsNfsInstance().Status.CapacityGb = state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb
	if qty, err := resource.ParseQuantity(fmt.Sprintf("%dGi", state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb)); err == nil {
		state.ObjAsNfsInstance().Status.Capacity = qty
	}

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		SuccessErrorNil().
		ErrorLogMessage("Error updating SAP NfsInstance status capacity").
		Run(ctx, state)
}
