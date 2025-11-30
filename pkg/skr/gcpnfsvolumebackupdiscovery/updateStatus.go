package gcpnfsvolumebackupdiscovery

import (
	"context"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	backupDiscovery := state.ObjAsGcpNfsVolumeBackupDiscovery()

	backupDiscovery.Status.DiscoverySnapshotTime = ptr.To(metav1.NewTime(time.Now()))
	backupDiscovery.Status.State = cloudresourcesv1beta1.JobStateDone
	backupDiscovery.Status.AvailableBackupsCount = ptr.To(len(state.backups))
	backupDiscovery.Status.AvailableBackupUris = make([]string, 0, len(state.backups))
	backupDiscovery.Status.AvailableBackups = make([]cloudresourcesv1beta1.AvailableBackup, 0, len(state.backups))
	for _, b := range state.backups {
		// Extract location_id/backup_id from the full backup name
		if uri := extractBackupUri(b.Name); uri != "" {
			backupDiscovery.Status.AvailableBackupUris = append(backupDiscovery.Status.AvailableBackupUris, uri)

			// Create AvailableBackup struct from backup labels
			availableBackup := cloudresourcesv1beta1.AvailableBackup{
				Uri:      uri,
				Location: extractBackupLocation(b.Name),
			}

			// Extract information from backup labels if available
			if b.Labels != nil {
				if shootName, ok := b.Labels[util.GcpLabelShootName]; ok {
					availableBackup.ShootName = shootName
				}
				if backupName, ok := b.Labels[util.GcpLabelSkrBackupName]; ok {
					availableBackup.BackupName = backupName
				}
				if backupNamespace, ok := b.Labels[util.GcpLabelSkrBackupNamespace]; ok {
					availableBackup.BackupNamespace = backupNamespace
				}
				if volumeName, ok := b.Labels[util.GcpLabelSkrVolumeName]; ok {
					availableBackup.VolumeName = volumeName
				}
				if volumeNamespace, ok := b.Labels[util.GcpLabelSkrVolumeNamespace]; ok {
					availableBackup.VolumeNamespace = volumeNamespace
				}
			}

			// Set creation time if available
			if b.CreateTime != "" {
				if creationTime, err := time.Parse(time.RFC3339, b.CreateTime); err == nil {
					availableBackup.CreationTime = ptr.To(metav1.NewTime(creationTime))
				}
			}

			backupDiscovery.Status.AvailableBackups = append(backupDiscovery.Status.AvailableBackups, availableBackup)
		}
	}

	return composed.UpdateStatus(backupDiscovery).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonReady,
			Message: "Successfully discovered available Nfs Volume Backups from GCP",
		}).
		ErrorLogMessage("Error: failed to set Done status on GcpNfsVolumeBackupDiscovery").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
