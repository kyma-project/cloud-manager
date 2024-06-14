package gcpnfsvolume

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadPersistentVolumeClaim(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	gcpNfsVolume := state.ObjAsGcpNfsVolume()

	pvc := &corev1.PersistentVolumeClaim{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      getVolumeClaimName(gcpNfsVolume),
	}, pvc)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error getting PersistentVolumeClaim by getVolumeName()", composed.StopWithRequeue, ctx)
	}

	if err != nil { // PVC not-found
		return nil, nil
	}

	pvcLabels := pvc.Labels
	parentGcpNfsVolumeName := pvcLabels[cloudresourcesv1beta1.LabelNfsVolName]

	if parentGcpNfsVolumeName != gcpNfsVolume.Name {
		errorMsg := fmt.Sprintf("Loaded PVC(%s/%s) belongs to another GcpNfsVolume (%s)", pvc.Namespace, pvc.Name, parentGcpNfsVolumeName)
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionPVCBelongsToAnotherGcpNfsVolume,
				Message: errorMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage(errorMsg).
			Run(ctx, state)
	}

	state.PVC = pvc

	return nil, nil
}
