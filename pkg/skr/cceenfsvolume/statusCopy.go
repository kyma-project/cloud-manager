package cceenfsvolume

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusCopy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpNfsInstance == nil {
		return nil, ctx
	}

	oldState := state.ObjAsCceeNfsVolume().Status.State

	changed, addedConditions, removedConditions, newState := composed.StatusCopyConditionsAndState(state.KcpNfsInstance, state.ObjAsCceeNfsVolume())
	changed = changed || state.ObjAsCceeNfsVolume().DeriveStateFromConditions()

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

	b := composed.PatchStatus(state.ObjAsCceeNfsVolume()).
		ErrorLogMessage("Error patching CceeNfsVolume status with conditions and state").
		FailedError(composed.StopWithRequeue)

	if meta.FindStatusCondition(state.ObjAsCceeNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeError) != nil {
		// KCP NfsInstance has error status
		// Stop the reconciliation
		b = b.
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("Forgetting CceeNfsVolume status with error condition")
	} else {
		// KCP NfsInstance is not in the error status, keep running
		b = b.SuccessErrorNil()
	}

	return b.Run(ctx, state)
}
