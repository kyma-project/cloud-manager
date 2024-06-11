package gcpnfsvolume

import (
	"context"
	"reflect"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
)

func modifyPersistentVolumeClaim(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

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
	}

	labels := getVolumeClaimLabels(nfsVolume)
	labelsChanged := !reflect.DeepEqual(state.PVC.Labels, labels)
	if labelsChanged {
		state.PVC.Labels = labels
	}

	annotations := getVolumeClaimAnnotations(nfsVolume)
	annotationsChanged := !reflect.DeepEqual(state.PVC.Annotations, annotations)
	if annotationsChanged {
		state.PVC.Annotations = annotations
	}

	if !(capacityChanged || labelsChanged || annotationsChanged) {
		return nil, nil
	}

	err := state.Cluster().K8sClient().Update(ctx, state.PVC)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating PersistentVolumeClaim", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
