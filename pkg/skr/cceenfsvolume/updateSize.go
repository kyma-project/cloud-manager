package cceenfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func updateSize(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsCceeNfsVolume().Spec.CapacityGb == state.KcpNfsInstance.Spec.Instance.OpenStack.SizeGb {
		return nil, ctx
	}

	newState := "Shrinking"
	if state.ObjAsCceeNfsVolume().Spec.CapacityGb > state.KcpNfsInstance.Spec.Instance.OpenStack.SizeGb {
		newState = "Extending"
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.
		WithValues(
			"oldCapacityGb", state.KcpNfsInstance.Spec.Instance.OpenStack.SizeGb,
			"newCapacityGb", state.ObjAsCceeNfsVolume().Spec.CapacityGb,
		).
		Info("Updating KCP NfsInstance capacity for CceeNfsInstance")

	p := map[string]interface{}{
		"spec": map[string]interface{}{
			"capacityGb": state.KcpNfsInstance.Spec.Instance.OpenStack.SizeGb,
		},
	}
	err := composed.PatchObj(ctx, state.ObjAsCceeNfsVolume(), p, state.KcpCluster.K8sClient())

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching KCP NfsInstance for CceeNfsVolume resize", composed.StopWithRequeue, ctx)
	}

	state.ObjAsCceeNfsVolume().Status.State = newState

	return composed.PatchStatus(state.ObjAsCceeNfsVolume()).
		ErrorLogMessage("Error patching CceeNfsVolume status with new status state for resize").
		// same as for success since we already updated kcp nfsInstance
		// in the new loop it will pick up the kcp state.state
		FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
		Run(ctx, state)
}
