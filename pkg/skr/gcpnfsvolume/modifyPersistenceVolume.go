package gcpnfsvolume

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func modifyPersistenceVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	//If GcpNfsVolume is marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR GcpNfsVolume is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}

	//Get GcpNfsVolume object
	nfsVolume := state.ObjAsGcpNfsVolume()
	capacity := gcpNfsVolumeCapacityToResourceQuantity(nfsVolume)

	//If GcpNfsVolume is not Ready state, continue.
	if !meta.IsStatusConditionTrue(nfsVolume.Status.Conditions, v1beta1.ConditionTypeReady) {
		return nil, nil
	}

	//If PV doesn't exist, continue.
	if state.PV == nil {
		return nil, nil
	}

	//Modify PV if any changes are done to GcpNfsVolume.
	changed := false
	if !capacity.Equal(state.PV.Spec.Capacity["storage"]) {
		changed = true
		state.PV.Spec.Capacity["storage"] = *capacity
		logger.Info("Detected modified PV capacity")
	}

	expectedLabels := getVolumeLabels(nfsVolume)
	if !areLabelsEqual(state.PV.Labels, expectedLabels) {
		changed = true
		state.PV.Labels = expectedLabels
		logger.Info("Detected modified PV labels")
	}

	expectedAnnotations := getVolumeAnnotations(nfsVolume)
	if !areAnnotationsSuperset(state.PV.Annotations, expectedAnnotations) {
		changed = true
		state.PV.Annotations = expectedAnnotations
		logger.Info("Detected desynced PV annotations")
	}

	//No changes to PV, continue.
	if !changed {
		return nil, nil
	}

	err := state.SkrCluster.K8sClient().Update(ctx, state.PV)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating PersistentVolume", composed.StopWithRequeue, ctx)
	}

	//continue
	return composed.StopWithRequeueDelay(1 * time.Second), nil
}
