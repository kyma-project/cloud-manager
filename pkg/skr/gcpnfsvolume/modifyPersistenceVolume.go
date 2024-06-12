package gcpnfsvolume

import (
	"context"
	"reflect"
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

	//If labels are different, update PV labels.
	labels := getVolumeLabels(nfsVolume)
	if !reflect.DeepEqual(state.PV.Labels, labels) {
		changed = true
		state.PV.Labels = labels
		logger.Info("Detected modified PV labels")
	}

	//If annotations are different, update PV annotations.
	annotations := getVolumeAnnotations(nfsVolume)
	if !reflect.DeepEqual(state.PV.Annotations, annotations) {
		changed = true
		state.PV.Annotations = annotations
		logger.Info("Detected modified PV annotations")
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
