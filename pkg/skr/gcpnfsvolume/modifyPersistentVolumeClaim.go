package gcpnfsvolume

import (
	"context"
	"reflect"
	"time"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
)

func modifyPersistentVolumeClaim(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//If GcpNfsVolume is marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR GcpNfsVolume is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}

	//Get GcpNfsVolume object
	nfsVolume := state.ObjAsGcpNfsVolume()
	capacity := resource.NewQuantity(int64(nfsVolume.Spec.CapacityGb)*1024*1024*1024, resource.BinarySI)

	//If GcpNfsVolume is not Ready state, continue.
	if !meta.IsStatusConditionTrue(nfsVolume.Status.Conditions, v1beta1.ConditionTypeReady) {
		return nil, nil
	}

	//If PVC doesn't exist, continue.
	if state.PVC == nil {
		return nil, nil
	}

	//Modify PVC if any changes are done to GcpNfsVolume.
	changed := false
	if !capacity.Equal(state.PVC.Spec.Resources.Requests["storage"]) {
		changed = true
		state.PVC.Spec.Resources.Requests["storage"] = *capacity
	}

	//If labels are different, update PVC labels.
	labels := getVolumeClaimLabels(nfsVolume, state)
	if !reflect.DeepEqual(state.PVC.Labels, labels) {
		changed = true
		state.PVC.Labels = labels
	}

	//If annotations are different, update PVC annotations.
	annotations := getVolumeClaimAnnotations(nfsVolume)
	if !reflect.DeepEqual(state.PVC.Annotations, annotations) {
		changed = true
		state.PVC.Annotations = annotations
	}

	//No changes to PVC, continue.
	if !changed {
		return nil, nil
	}

	err := state.SkrCluster.K8sClient().Update(ctx, state.PVC)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating PersistentVolumeClaim", composed.StopWithRequeue, ctx)
	}

	//continue
	return composed.StopWithRequeueDelay(1 * time.Second), nil
}
