package gcpnfsvolume

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
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

	labelsBuilder := util.NewLabelBuilder()
	labelsBuilder.WithCustomLabels(getVolumeClaimLabels(state.ObjAsGcpNfsVolume()))
	labelsBuilder.WithCloudManagerDefaults()

	storage := state.PV.Spec.Capacity["storage"]
	labelsBuilder.WithCustomLabel(cloudresourcesv1beta1.LabelStorageCapacity, storage.String())

	pvcLabels := labelsBuilder.Build()

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   state.Obj().GetNamespace(),
			Name:        getVolumeClaimName(state.ObjAsGcpNfsVolume()),
			Labels:      pvcLabels,
			Annotations: getVolumeClaimAnnotations(state.ObjAsGcpNfsVolume()),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:  state.PV.GetName(), // connection to PV
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storage,
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
