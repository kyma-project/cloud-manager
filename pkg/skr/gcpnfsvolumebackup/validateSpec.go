package gcpnfsvolumebackup

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateLocation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}
	location := state.ObjAsGcpNfsVolumeBackup().Spec.Location
	if location == "" {
		if feature.GcpNfsVolumeAutomaticLocationAllocation.Value(ctx) {
			return nil, nil
		}
		state.ObjAsGcpNfsVolumeBackup().Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		return composed.UpdateStatus(state.ObjAsGcpNfsVolumeBackup()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonLocationInvalid,
				Message: "Location is required",
			}).
			ErrorLogMessage("Error updating GcpNfsVolumeBackup status with invalid location").
			Run(ctx, state)
	} else {
		// if validation succeeds, we don't need to update the status
		return nil, nil
	}
}
