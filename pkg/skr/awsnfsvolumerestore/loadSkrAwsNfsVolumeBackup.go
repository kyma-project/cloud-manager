package awsnfsvolumerestore

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadSkrAwsNfsVolumeBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsAwsNfsVolumeRestore()

	//If the object is being deleted continue...
	if composed.IsMarkedForDeletion(restore) {
		return nil, nil
	}
	backupName := restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace())

	skrAwsNfsVolumeBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
	err := state.Cluster().K8sClient().Get(
		ctx,
		backupName,
		skrAwsNfsVolumeBackup,
	)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading SKR AwsNfsVolumeBackup", composed.StopWithRequeue, ctx)
	}

	if err == nil {
		state.skrAwsNfsVolumeBackup = skrAwsNfsVolumeBackup
		return nil, nil
	}

	// skrAwsNfsVolumeBackup does not exist
	// * update state to error
	// * set error condition
	// * stop and forget
	restore.SetState(cloudresourcesv1beta1.JobStateError)
	return composed.PatchStatus(restore).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolumeBackup,
			Message: fmt.Sprintf("AwsNfsVolumeBackup %s does not exist", backupName),
		}).
		ErrorLogMessage("Failed updating AwsNfsVolumeRestore status with error state due to missing AwsNfsVolumeBackup").
		SuccessLogMsg("Forgetting AwsNfsVolumeRestore with missing AwsNfsVolumeBackup").
		FailedError(composed.StopAndForget).
		Run(ctx, state)
}
