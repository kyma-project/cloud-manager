package gcpnfsvolume

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Validate that if a PVC with expected name exists, it belongs to current GCPNfsVolume.
func validatePVC(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	nfsVolume := state.ObjAsGcpNfsVolume()

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	pvcName := getVolumeClaimName(nfsVolume)
	pvc := &corev1.PersistentVolumeClaim{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{Name: pvcName, Namespace: nfsVolume.Namespace}, pvc)

	if apierrors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting PersistentVolumeClaim by name", composed.StopWithRequeue, ctx)
	}

	parentName, nameLabelExists := pvc.Labels[cloudresourcesv1beta1.LabelNfsVolName]
	parentNamespace, namespaceLabelExists := pvc.Labels[cloudresourcesv1beta1.LabelNfsVolNS]
	if nameLabelExists && namespaceLabelExists && parentName == nfsVolume.Name && parentNamespace == nfsVolume.Namespace {
		return nil, nil
	}

	nfsVolume.Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
	errorMsg := fmt.Sprintf("Loaded PVC(%s/%s) belongs to another GcpNfsVolume (%s/%s)", pvc.Namespace, pvc.Name, parentNamespace, parentName)

	return composed.UpdateStatus(nfsVolume).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonPVCNameInvalid,
			Message: errorMsg,
		}).
		RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
		ErrorLogMessage(errorMsg).
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
