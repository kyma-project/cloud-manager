package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func modifyPersistentVolumeClaim(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR GcpNfsVolume is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}

	nfsVolume := state.ObjAsGcpNfsVolume()
	capacity := gcpNfsVolumeCapacityToResourceQuantity(nfsVolume)

	if !meta.IsStatusConditionTrue(nfsVolume.Status.Conditions, v1beta1.ConditionTypeReady) {
		return nil, nil
	}

	if state.PVC == nil {
		return nil, nil
	}

	capacityChanged := !capacity.Equal(state.PVC.Spec.Resources.Requests["storage"])
	if capacityChanged {
		state.PVC.Spec.Resources.Requests["storage"] = *capacity
		logger.Info("Detected modified PVC capacity")
	}

	expectedLabels := getVolumeClaimLabels(nfsVolume)
	labelsChanged := !areLabelsEqual(state.PVC.Labels, expectedLabels)
	if labelsChanged {
		state.PVC.Labels = expectedLabels
		logger.Info("Detected modified PVC labels")
	}

	expectedAnnotations := getVolumeClaimAnnotations(nfsVolume)
	annotationsDesynced := !areAnnotationsSuperset(state.PVC.Annotations, expectedAnnotations) // PVC controller will keep adding "pv.kubernetes.io/bind-completed=yes" annotation, so we must check if we are actual is superset of expected
	if annotationsDesynced {
		state.PVC.Annotations = expectedAnnotations
		logger.Info("Detected desynced PVC annotations")
	}

	if !(capacityChanged || labelsChanged || annotationsDesynced) {
		return nil, nil
	}

	err := state.Cluster().K8sClient().Update(ctx, state.PVC)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating PersistentVolumeClaim", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
