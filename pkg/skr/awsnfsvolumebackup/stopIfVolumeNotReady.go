package awsnfsvolumebackup

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func stopIfVolumeNotReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	isReady := meta.IsStatusConditionTrue(state.ObjAsAwsNfsVolumeBackup().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if isReady {
		return nil, nil
	}

	state.ObjAsAwsNfsVolumeBackup().SetState(cloudresourcesv1beta1.StateError)

	return composed.UpdateStatus(state.ObjAsAwsNfsVolumeBackup()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonNfsVolumeNotReady,
			Message: fmt.Sprintf("The AwsNfsVolume is not ready"),
		}).
		ErrorLogMessage("Failed updating AwsNfsVolumeBackup error status with NfsVolumeNotReady condition").
		SuccessLogMsg("Forgeting AwsNfsVolumeBackup with NfsVolume not ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
