package sapnfsvolume

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
	if composed.IsMarkedForDeletion(state.ObjAsSapNfsVolume()) {
		return nil, ctx
	}
	if composed.IsMarkedForDeletion(state.KcpNfsInstance) {
		return nil, ctx
	}

	oldState := state.ObjAsSapNfsVolume().Status.State

	changed, addedConditions, removedConditions, newState := composed.StatusCopyConditionsAndState(state.KcpNfsInstance, state.ObjAsSapNfsVolume())
	changed = changed || state.ObjAsSapNfsVolume().DeriveStateFromConditions()

	if state.ObjAsSapNfsVolume().Status.Capacity != state.KcpNfsInstance.Status.Capacity {
		changed = true
		state.ObjAsSapNfsVolume().Status.Capacity = state.KcpNfsInstance.Status.Capacity
	}

	if !changed {
		return nil, ctx
	}

	logger.
		WithValues(
			"allConditions", pie.Map(state.ObjAsSapNfsVolume().Status.Conditions, func(c metav1.Condition) string {
				return fmt.Sprintf("%s/%s/%s", c.Type, c.Reason, c.Message)
			}),
			"addedConditions", addedConditions,
			"removedConditions", removedConditions,
			"newState", newState,
			"oldState", oldState,
		).
		Info("Updating SapNfsVolume status with conditions and state")

	b := composed.PatchStatus(state.ObjAsSapNfsVolume()).
		ErrorLogMessage("Error patching SapNfsVolume status with conditions and state").
		FailedError(composed.StopWithRequeue)

	if meta.FindStatusCondition(state.ObjAsSapNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeError) != nil {
		// KCP NfsInstance has error status
		// Stop the reconciliation
		b = b.
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("Forgetting SapNfsVolume status with error condition")
	} else {
		// KCP NfsInstance is not in the error status, keep running
		b = b.SuccessErrorNil()
	}

	return b.Run(ctx, state)
}
