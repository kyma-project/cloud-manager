package awsnfsvolumebackup

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadKcpAwsNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	backup := state.ObjAsAwsNfsVolumeBackup()

	//If the object is being deleted continue...
	if composed.IsMarkedForDeletion(backup) {
		return nil, nil
	}

	kcpNfsInstance := &cloudcontrolv1beta1.NfsInstance{}
	err := state.KcpCluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef().Namespace,
		Name:      state.skrAwsNfsVolume.Status.Id,
	}, kcpNfsInstance)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP NfsInstance", composed.StopWithRequeue, ctx)
	}

	if err == nil {
		state.kcpAwsNfsInstance = kcpNfsInstance
		return nil, nil
	}

	// kcpNfsInstance does not exist
	// * update state to error
	// * set error condition
	// * stop and forget
	state.ObjAsAwsNfsVolumeBackup().SetState(cloudresourcesv1beta1.StateError)
	return composed.PatchStatus(state.ObjAsAwsNfsVolumeBackup()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolume,
			Message: fmt.Sprintf("NfsInstance %s does not exist", state.skrAwsNfsVolume.Status.Id),
		}).
		ErrorLogMessage("Failed updating AwsNfsVolumeBackup status with error state due to missing NfsInstance").
		SuccessLogMsg("Forgetting AwsNfsVolumeBackup with missing NfsInstance").
		FailedError(composed.StopAndForget).
		Run(ctx, state)
}
