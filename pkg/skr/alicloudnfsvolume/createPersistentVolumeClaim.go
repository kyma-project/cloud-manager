package alicloudnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func createPersistentVolumeClaim(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}
	if state.PVC != nil {
		logger.Info("PersistentVolumeClaim for AlicloudNfsVolume already exists")
		return nil, nil
	}

	if state.Volume == nil {
		return composed.StopWithRequeueDelay(2 * util.Timing.T100ms()), nil
	}

	pvc := &corev1.PersistentVolumeClaim{
		Namespace:   state.Obj().GetNamespace(),
		Name:        getVolumeClaimName(state.ObjAsAlicloudNfsVolume()),
		Labels:      getVolumeClaimLabels(state.ObjAsAlicloudNfsVolume()),
		Annotations: getVolumeClaimAnnotations(state.ObjAsAlicloudNfsVolume()),
		Finalizers: []string{
			api.CommonFinalizerDeletionHook,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:  state.Volume.GetName(), // connection to PV
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": state.ObjAsAlicloudNfsVolume().Spec.Capacity,
				},
			},
			StorageClassName: new(""),
			VolumeMode:       ptr.To(corev1.PersistentVolumeFilesystem),
		},
	}
	err := state.Cluster().K8sClient().Create(ctx, pvc)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating PVC for PV", composed.StopWithRequeue, ctx)
	}

	logger.Info("PVC for AliCloud PV created")

	return nil, nil
}
