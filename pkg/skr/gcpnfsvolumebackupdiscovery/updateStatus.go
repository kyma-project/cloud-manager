package gcpnfsvolumebackupdiscovery

import (
	"context"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	backupDiscovery := state.ObjAsGcpNfsVolumeBackupDiscovery()

	backupDiscovery.Status.DiscoverySnapshotTime = ptr.To(metav1.NewTime(time.Now()))
	backupDiscovery.Status.State = cloudresourcesv1beta1.JobStateDone
	backupDiscovery.Status.AvailableBackupsCount = ptr.To(len(state.backups))
	backupDiscovery.Status.AvailableBackupUris = make([]string, 0, len(state.backups))
	for _, b := range state.backups {
		backupDiscovery.Status.AvailableBackupUris = append(backupDiscovery.Status.AvailableBackupUris, b.Name)
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
