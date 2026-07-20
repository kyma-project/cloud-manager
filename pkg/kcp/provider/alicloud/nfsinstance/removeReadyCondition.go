package nfsinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

// removeReadyCondition clears the Ready condition at the start of deletion.
func removeReadyCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !meta.IsStatusConditionTrue(*state.ObjAsNfsInstance().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		return nil, ctx
	}

	meta.RemoveStatusCondition(state.ObjAsNfsInstance().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)

	return composed.UpdateStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error removing Ready condition from AliCloud KCP NfsInstance").
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
