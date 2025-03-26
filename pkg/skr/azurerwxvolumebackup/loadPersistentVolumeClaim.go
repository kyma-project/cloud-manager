package azurerwxvolumebackup

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	client2 "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadPersistentVolumeClaim(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAzureRwxVolumeBackup()

	pvc := &corev1.PersistentVolumeClaim{}
	err := state.Cluster().K8sClient().Get(ctx, backup.Spec.Source.Pvc.ToNamespacedName(backup.Namespace), pvc)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error in loading PersistentVolumeClaim", composed.StopWithRequeue, ctx)
	}

	if err != nil {
		backup.Status.State = cloudresourcesv1beta1.JobStateFailed
		logger.Error(err, "PersistentVolumeClaim was not found.", "PVC", backup.Spec.Source.Pvc.ToNamespacedName(backup.Namespace))
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonPvcNotFound,
				Message: "Specified destination was not found",
			}).
			ErrorLogMessage("Error patching AzureRwxVolumeBackup status").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if pvc.Status.Phase != "Bound" {
		backup.Status.State = cloudresourcesv1beta1.JobStateFailed
		logger.Error(nil, "Specified destination PVC is not in 'Bound' state", "PVC", backup.Spec.Source.Pvc.ToNamespacedName(backup.Namespace), "Phase", pvc.Status.Phase)
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonPvcNotBound,
				Message: fmt.Sprintf("Specified destination PVC is in invalid state %v", pvc.Status.Phase),
			}).
			ErrorLogMessage("Error patching AzureRwxVolumeBackup status").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if !client2.IsPvcProvisionerAzureCsiDriver(pvc.Annotations) {
		logger.Error(nil, "Specified destination PVC is not provisioned by Azure CSI driver (file.csi.azure.com)", "PVC", backup.Spec.Source.Pvc.ToNamespacedName(backup.Namespace))
		backup.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonInvalidProvisioner,
				Message: "Specified destination PVC is not provisioned by Azure CSI driver (file.csi.azure.com)",
			}).
			ErrorLogMessage("Error patching AzureRwxVolumeBackup status").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	return nil, nil

}
