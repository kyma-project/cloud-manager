package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createPersistentVolumeClaim(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.PVC != nil {
		logger.Info("PersistentVolumeClaim for GcpNfsVolume already exists")
		return nil, nil
	}

	if state.PV == nil {
		return nil, nil
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   state.Obj().GetNamespace(),
			Name:        getVolumeClaimName(state.ObjAsGcpNfsVolume()),
			Labels:      getVolumeClaimLabels(state.ObjAsGcpNfsVolume(), state),
			Annotations: getVolumeClaimAnnotations(state.ObjAsGcpNfsVolume()),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:  state.PV.GetName(), // connection to PV
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": state.PV.Spec.Capacity["storage"],
				},
			},
			StorageClassName: ptr.To(""),
			VolumeMode:       ptr.To(corev1.PersistentVolumeFilesystem),
		},
	}
	err := state.Cluster().K8sClient().Create(ctx, pvc)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating PVC for PV", composed.StopWithRequeue, ctx)
	}

	logger.Info("PVC for Gcp PV created")

	return nil, nil
}