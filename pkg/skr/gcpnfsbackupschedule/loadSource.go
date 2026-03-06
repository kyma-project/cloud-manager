package gcpnfsbackupschedule

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func loadSource(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	namespace := schedule.GetSourceRef().Namespace
	if len(namespace) == 0 {
		namespace = schedule.GetNamespace()
	}

	source := &cloudresourcesv1beta1.GcpNfsVolume{}
	err := state.SkrCluster.K8sClient().Get(ctx, types.NamespacedName{
		Name:      schedule.GetSourceRef().Name,
		Namespace: namespace,
	}, source)

	if err != nil {
		schedule.SetState(cloudresourcesv1beta1.StateError)
		logger.Error(err, "Error getting SourceRef")
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonNfsVolumeNotFound,
				Message: "Error loading SourceRef",
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	// Check if the source has a ready condition
	volumeReady := meta.FindStatusCondition(*source.Conditions(), cloudresourcesv1beta1.ConditionTypeReady)

	if volumeReady == nil || volumeReady.Status != metav1.ConditionTrue {
		logger.Info("SourceRef is not ready")
		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonNfsVolumeNotReady,
				Message: "SourceRef is not ready",
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessLogMsg("Error getting SourceRef").
			Run(ctx, state)
	}

	state.Source = source

	return nil, nil
}
