package cceenfsvolume

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusCopy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	oldState := state.ObjAsCceeNfsVolume().Status.State

	changed, addedConditions, removedConditions, newState := composed.StatusCopyConditionsAndState(state.KcpNfsInstance, state.ObjAsCceeNfsVolume())
	if !changed {
		return nil, ctx
	}

	logger.
		WithValues(
			"allConditions", pie.Map(state.ObjAsCceeNfsVolume().Status.Conditions, func(c metav1.Condition) string {
				return fmt.Sprintf("%s/%s/%s", c.Type, c.Reason, c.Message)
			}),
			"addedConditions", addedConditions,
			"removedConditions", removedConditions,
			"newState", newState,
			"oldState", oldState,
		).
		Info("Updating CceeNfsVolume status with conditions and state")

	return composed.PatchStatus(state.ObjAsCceeNfsVolume()).
		ErrorLogMessage("Error patching CceeNfsVolume status with conditions and state").
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
