package sapnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func pvcCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}
	// if PV is not created then PVC can not be created
	if state.PV == nil {
		return nil, ctx
	}
	// if PVC is already created, nothing to do
	if state.PVC != nil {
		return nil, ctx
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: state.ObjAsSapNfsVolume().GetNamespace(),
			Name:      state.ObjAsSapNfsVolume().GetPVCName(),
			Labels: util.NewLabelBuilder().
				WithCustomLabels(state.ObjAsSapNfsVolume().GetPVCLabels()).
				WithCloudManagerDefaults().
				Build(),
			Annotations: state.ObjAsSapNfsVolume().GetPVCAnnotations(),
			Finalizers: []string{
				api.CommonFinalizerDeletionHook,
			},
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
		return composed.LogErrorAndReturn(err, "Error creating PVC for SapNfsVolume", composed.StopWithRequeue, ctx)
	}

	logger.
		WithValues("pvcName", pvc.Name).
		Info("Created PVC for SapNfsVolume")

	state.PVC = pvc

	return nil, ctx
}
