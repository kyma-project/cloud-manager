package v2

import (
	"context"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	backup := state.ObjAsGcpNfsVolumeBackup()
	logger := composed.LoggerFromCtx(ctx)

	// Check if the backup has a ready condition
	backupReady := meta.FindStatusCondition(backup.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	// If file backup is ready, update the status of the backup
	needsUpdate := false

	// Update Ready condition if not set
	if backupReady == nil || backupReady.Status != metav1.ConditionTrue {
		logger.Info("Updating SKR GcpNfsVolumeBackup status with Ready condition")
		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupReady
		needsUpdate = true
	}

	// Update capacity if it's time
	if state.isTimeForCapacityUpdate() {
		logger.Info("Updating SKR GCPNfsVolumeBackup status with Capacity")
		size := state.fileBackup.StorageBytes
		capacity := resource.NewQuantity(size, resource.BinarySI)
		backup.Status.Capacity = *capacity
		backup.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now().UTC()}
		needsUpdate = true
	}

	// Mirror labels to status
	if !state.HasAllStatusLabels() {
		backup.Status.FileStoreBackupLabels = state.fileBackup.Labels
		needsUpdate = true
	}

	if needsUpdate {
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: "Backup is ready for use.",
			}).
			SuccessLogMsg("GcpNfsVolumeBackup status updated").
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, nil
}
