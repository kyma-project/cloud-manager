package sapnfsvolumesnapshotrestore

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func restoreNewVolumeWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	// Reload the created volume to get current status
	vol := &cloudresourcesv1beta1.SapNfsVolume{}
	err := state.SkrCluster.K8sClient().Get(ctx, types.NamespacedName{
		Name:      state.CreatedVolume.Name,
		Namespace: state.CreatedVolume.Namespace,
	}, vol)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading created SapNfsVolume", composed.StopWithRequeue, ctx)
	}

	state.CreatedVolume = vol

	// Check if volume is Ready
	volumeReady := meta.FindStatusCondition(vol.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
	if volumeReady != nil && volumeReady.Status == metav1.ConditionTrue {
		// Volume is ready, record in status
		restore.Status.CreatedVolume = &corev1.ObjectReference{
			Kind:       "SapNfsVolume",
			APIVersion: cloudresourcesv1beta1.GroupVersion.String(),
			Name:       vol.Name,
			Namespace:  vol.Namespace,
		}
		return nil, ctx
	}

	// Check if volume entered an error state
	if vol.Status.State == cloudresourcesv1beta1.StateError || vol.Status.State == cloudresourcesv1beta1.StateFailed {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("created SapNfsVolume %s/%s is in %s state", vol.Namespace, vol.Name, vol.Status.State), "New volume failed")
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsRestoreFailed,
				Message: fmt.Sprintf("Created SapNfsVolume %s/%s entered %s state", vol.Namespace, vol.Name, vol.Status.State),
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	// Volume is still being created, requeue
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
