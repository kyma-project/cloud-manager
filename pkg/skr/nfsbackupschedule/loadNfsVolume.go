package nfsbackupschedule

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func loadNfsVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsSchedule()
	logger := composed.LoggerFromCtx(ctx)

	//If marked for deletion, return
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	logger.WithValues("Nfs Backup Schedule :", schedule.GetName()).Info("Loading NfsVolume")

	//Load the nfsVolume object
	nfsVolume, err := getProviderSpecificNfsObject(ctx, state)
	if err != nil {
		schedule.SetState(cloudresourcesv1beta1.StateError)
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonNfsVolumeNotFound,
				Message: "Error loading NfsVolume",
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error getting NfsVolume").
			Run(ctx, state)
	}

	//Check if the nfsVolume has a ready condition
	volumeReady := meta.FindStatusCondition(*nfsVolume.Conditions(), cloudresourcesv1beta1.ConditionTypeReady)

	//If the nfsVolume is not ready, return an error
	if volumeReady == nil || volumeReady.Status != metav1.ConditionTrue {
		logger.WithValues("NfsBackupSchedule", schedule.GetName()).Info("NfsVolume is not ready")
		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonNfsVolumeNotReady,
				Message: "NfsVolume is not ready",
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg("Error getting NfsVolume").
			Run(ctx, state)
	}

	//Store the NfsVolume in state
	state.NfsVolume = nfsVolume

	return nil, nil
}

func getProviderSpecificNfsObject(ctx context.Context, state *State) (composed.ObjWithConditions, error) {
	schedule := state.ObjAsSchedule()
	key := types.NamespacedName{
		Name:      schedule.GetSourceRef().Name,
		Namespace: schedule.GetSourceRef().Namespace,
	}
	var nfsVolume composed.ObjWithConditions

	switch state.Scope.Spec.Provider {
	case cloudcontrolv1beta1.ProviderGCP:
		nfsVolume = &cloudresourcesv1beta1.GcpNfsVolume{}
	case cloudcontrolv1beta1.ProviderAws:
		nfsVolume = &cloudresourcesv1beta1.AwsNfsVolume{}
	default:
		return nil, fmt.Errorf("provider %s not supported", state.Scope.Spec.Provider)
	}
	err := state.SkrCluster.K8sClient().Get(ctx, key, nfsVolume)
	return nfsVolume, err
}
