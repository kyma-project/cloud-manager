package backupschedule

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

	//If marked for deletion, return
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	logger.WithValues("NfsBackupBackupSchedule", schedule.GetName()).Info("Loading SourceRef")

	//Load the sourceRef object
	sourceRef, err := getSourceRef(ctx, state)
	if err != nil {
		schedule.SetState(cloudresourcesv1beta1.StateError)
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonNfsVolumeNotFound,
				Message: "Error loading SourceRef",
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessLogMsg("Error getting SourceRef").
			Run(ctx, state)
	}

	//Check if the sourceRef has a ready condition
	volumeReady := meta.FindStatusCondition(*sourceRef.Conditions(), cloudresourcesv1beta1.ConditionTypeReady)

	//If the sourceRef is not ready, return an error
	if volumeReady == nil || volumeReady.Status != metav1.ConditionTrue {
		logger.WithValues("GcpNfsBackupSchedule", schedule.GetName()).Info("SourceRef is not ready")
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

	//Store the SourceRef in state
	state.SourceRef = sourceRef

	return nil, nil
}

func getSourceRef(ctx context.Context, state *State) (composed.ObjWithConditions, error) {
	schedule := state.ObjAsBackupSchedule()
	namespace := schedule.GetSourceRef().Namespace
	if len(namespace) <= 0 {
		namespace = schedule.GetNamespace()
	}

	key := types.NamespacedName{
		Name:      schedule.GetSourceRef().Name,
		Namespace: namespace,
	}

	source := state.backupImpl.emptySourceObject()
	err := state.SkrCluster.K8sClient().Get(ctx, key, source)
	return source, err
}
