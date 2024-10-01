package awsnfsvolumerestore

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func stopIfBackupNotReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsAwsNfsVolumeRestore()

	//If the object is being deleted continue...
	if composed.IsMarkedForDeletion(restore) {
		return nil, nil
	}

	isReady := meta.IsStatusConditionTrue(state.skrAwsNfsVolumeBackup.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if isReady {
		return nil, nil
	}

	restore.SetState(cloudresourcesv1beta1.JobStateError)

	return composed.PatchStatus(restore).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeBackupNotReady,
			Message: "The AwsNfsVolume is not ready",
		}).
		ErrorLogMessage("Failed updating AwsNfsVolumeRestore error status with NfsVolumeBackupNotReady condition").
		SuccessLogMsg("Forgetting AwsNfsVolumeRestore with NfsVolumeBackup not ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
