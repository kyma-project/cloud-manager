package awsnfsvolumebackup

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadSkrAwsNfsVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	name := state.ObjAsAwsNfsVolumeBackup().Spec.Source.Volume.ToNamespacedName(state.Obj().GetNamespace())

	skrAwsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
	err := state.Cluster().K8sClient().Get(
		ctx,
		name,
		skrAwsNfsVolume,
	)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading SKR AwsNfsVolume", composed.StopWithRequeue, ctx)
	}

	if err == nil {
		state.skrAwsNfsVolume = skrAwsNfsVolume
		return nil, nil
	}

	// skrAwsNfsVolume does not exist
	// * update state to error
	// * set error condition
	// * stop and forget
	state.ObjAsAwsNfsVolumeBackup().SetState(cloudresourcesv1beta1.StateError)
	return composed.UpdateStatus(state.ObjAsAwsNfsVolumeBackup()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonMissingNfsVolume,
			Message: fmt.Sprintf("AwsNfsVolume %s does not exist", name),
		}).
		ErrorLogMessage("Failed updating AwsNfsVolumeBackup status with error state due to missing AwsNfsVolume").
		SuccessLogMsg("Forgetting AwsNfsVolumeBackup with missing AwsNfsVolume").
		FailedError(composed.StopAndForget).
		Run(ctx, state)
}
