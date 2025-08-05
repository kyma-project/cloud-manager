package gcpnfsvolumebackup

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateCapacity(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	backup := state.ObjAsGcpNfsVolumeBackup()
	logger := composed.LoggerFromCtx(ctx)

	// If deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If capacityUpdate is not due yet, continue (?)
	if !state.isTimeForCapacityUpdate() {
		logger.Info("Not yet time for updating capacity, continuing...")
		return nil, nil
	}

	// Update Capacity and timestamp (?)
	logger.Info("Updating SKR GCPNfsVolumeBackup status with Capacity")
	size := state.fileBackup.CapacityGb
	capacity := resource.NewQuantity(size, resource.BinarySI)
	backup.Status.Capacity = *capacity
	backup.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now().UTC()}

	return composed.PatchStatus(backup).
		SuccessErrorNil().
		Run(ctx, state)

}
