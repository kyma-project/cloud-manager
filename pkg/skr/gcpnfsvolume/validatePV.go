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

func validatePV(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	nfsVolume := state.ObjAsGcpNfsVolume()

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	pvName := getVolumeName(nfsVolume)
	pv := &corev1.PersistentVolume{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, pv)

	if apierrors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting PersistentVolume by name", composed.StopWithRequeue, ctx)
	}

	parentName, parentNameExists := pv.Labels[cloudresourcesv1beta1.LabelNfsVolName]
	parentNamespace, parentNamespaceExists := pv.Labels[cloudresourcesv1beta1.LabelNfsVolNS]
	if parentNameExists && parentNamespaceExists && parentName == nfsVolume.Name && parentNamespace == nfsVolume.Namespace {
		return nil, nil
	}

	nfsVolume.Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
	errorMsg := fmt.Sprintf("Desired PV(%s) already exists with different owner", pvName)

	return composed.PatchStatus(nfsVolume).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonPVNameInvalid,
			Message: errorMsg,
		}).
		RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
		ErrorLogMessage(errorMsg).
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
