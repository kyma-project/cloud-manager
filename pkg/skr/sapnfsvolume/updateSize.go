package sapnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func updateSize(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.IsMarkedForDeletion(state.ObjAsSapNfsVolume()) {
		return nil, ctx
	}
	if state.KcpNfsInstance == nil {
		return nil, ctx
	}

	if state.ObjAsSapNfsVolume().Spec.CapacityGb == state.KcpNfsInstance.Spec.Instance.OpenStack.SizeGb {
		return nil, ctx
	}

	newState := "Shrinking"
	if state.ObjAsSapNfsVolume().Spec.CapacityGb > state.KcpNfsInstance.Spec.Instance.OpenStack.SizeGb {
		newState = "Extending"
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.
		WithValues(
			"oldCapacityGb", state.KcpNfsInstance.Spec.Instance.OpenStack.SizeGb,
			"newCapacityGb", state.ObjAsSapNfsVolume().Spec.CapacityGb,
		).
		Info("Updating KCP NfsInstance capacity for SapNfsVolume")

	p := map[string]interface{}{
		"spec": map[string]interface{}{
			"capacityGb": state.KcpNfsInstance.Spec.Instance.OpenStack.SizeGb,
		},
	}
	err := composed.MergePatchObj(ctx, state.ObjAsSapNfsVolume(), p, state.KcpCluster.K8sClient())

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching KCP NfsInstance for SapNfsVolume resize", composed.StopWithRequeue, ctx)
	}

	state.ObjAsSapNfsVolume().Status.State = newState

	return composed.PatchStatus(state.ObjAsSapNfsVolume()).
		ErrorLogMessage("Error patching SapNfsVolume status with new status state for resize").
		// same as for success since we already updated kcp nfsInstance
		// in the new loop it will pick up the kcp state.state
		FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
		Run(ctx, state)
}
