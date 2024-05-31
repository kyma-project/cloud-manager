package awsnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createPersistentVolumeClaim(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}
	if state.PVC != nil {
		logger.Info("PersistentVolumeClaim for AwsNfsVolume already exists")
		return nil, nil
	}

	if state.Volume == nil {
		return composed.StopWithRequeueDelay(2 * util.Timing.T100ms()), nil
	}

	//lbls := map[string]string{
	//	"whatever-label": "additional-lebels-from-the-ticket",
	//	"storageGB":      state.ObjAsAwsNfsVolume().Spec.Capacity.String(),
	//}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: state.Obj().GetNamespace(),
			Name:      getVolumeName(state.ObjAsAwsNfsVolume()),
			// Labels:    lbls,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:  state.Volume.GetName(), // connection to PV
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": state.ObjAsAwsNfsVolume().Spec.Capacity,
				},
			},
			StorageClassName: func() *string { v := "standard"; return &v }(),                                             // Set the desired StorageClass
			VolumeMode:       func() *corev1.PersistentVolumeMode { v := corev1.PersistentVolumeFilesystem; return &v }(), // Set the VolumeMode

		},
	}
	err := state.Cluster().K8sClient().Create(ctx, pvc)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating PVC for PV", composed.StopWithRequeue, ctx)
	}

	logger.Info("PVC for AWS PV created")

	return nil, nil
}
