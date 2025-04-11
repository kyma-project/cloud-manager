package azurerwxvolumerestore

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadPersistentVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	azureRwxVolumeRestore := state.ObjAsAzureRwxVolumeRestore()
	pvName := state.pvc.Spec.VolumeName
	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Loading PersistentVolume", "PV", pvName)

	pv := &corev1.PersistentVolume{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, pv)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error in loading PersistentVolume", composed.StopWithRequeue, ctx)
	}

	if err != nil {
		azureRwxVolumeRestore.Status.State = cloudresourcesv1beta1.JobStateFailed
		logger.Error(err, "PersistentVolume was not found.", "PV", pvName)
		return composed.PatchStatus(azureRwxVolumeRestore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonPvNotFound,
				Message: "Persistent volume was not found",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if pv.Status.Phase != "Bound" {
		logger.Error(nil, "PV for specified destination PVC is not in 'Bound' state", "PV", pvName, "Phase", pv.Status.Phase)
		azureRwxVolumeRestore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(azureRwxVolumeRestore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonPvNotBound,
				Message: fmt.Sprintf("PV for specified destination PVC is in invalid state %v", pv.Status.Phase),
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}
	resourceGroupName, storageAccountName, fileShareName, _, _, err := azurerwxvolumebackupclient.ParsePvVolumeHandle(pv.Spec.CSI.VolumeHandle)
	if err != nil {
		azureRwxVolumeRestore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(azureRwxVolumeRestore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonInvalidVolumeHandle,
				Message: fmt.Sprintf("Persistant Volume has an unexpected volume handle: %v", pv.Spec.CSI.VolumeHandle),
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}
	state.resourceGroupName = resourceGroupName
	state.storageAccountName = storageAccountName
	state.fileShareName = fileShareName
	return nil, nil

}
