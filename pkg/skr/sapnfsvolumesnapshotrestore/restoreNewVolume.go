package sapnfsvolumesnapshotrestore

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func restoreNewVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	newVolTemplate := restore.Spec.Destination.NewVolume
	ns := newVolTemplate.Metadata.Namespace
	if ns == "" {
		ns = restore.Namespace
	}

	// Check if the volume already exists (idempotency)
	existing := &cloudresourcesv1beta1.SapNfsVolume{}
	err := state.SkrCluster.K8sClient().Get(ctx, types.NamespacedName{
		Name:      newVolTemplate.Metadata.Name,
		Namespace: ns,
	}, existing)
	if err == nil {
		// Volume already exists, store it and proceed to wait
		state.CreatedVolume = existing
		return nil, ctx
	}

	// Create the new SapNfsVolume with the snapshot-id annotation
	newVolume := &cloudresourcesv1beta1.SapNfsVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SapNfsVolume",
			APIVersion: cloudresourcesv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      newVolTemplate.Metadata.Name,
			Namespace: ns,
			Labels:    newVolTemplate.Metadata.Labels,
			Annotations: mergeAnnotations(
				newVolTemplate.Metadata.Annotations,
				map[string]string{
					cloudresourcesv1beta1.AnnotationSnapshotId: state.SourceSnapshot.Status.OpenstackId,
				},
			),
		},
		Spec: newVolTemplate.Spec,
	}

	err = state.SkrCluster.K8sClient().Create(ctx, newVolume)
	if err != nil {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Error creating new SapNfsVolume from snapshot")
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsRestoreFailed,
				Message: fmt.Sprintf("Error creating new SapNfsVolume: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	state.CreatedVolume = newVolume

	return nil, ctx
}

func mergeAnnotations(base map[string]string, extra map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		result[k] = v
	}
	return result
}
