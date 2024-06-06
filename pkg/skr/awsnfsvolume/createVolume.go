package awsnfsvolume

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}
	if state.Volume != nil {
		logger.Info("PersistentVolume for AwsNfsVolume already exists")
		return nil, nil
	}

	kcpCondReady := meta.FindStatusCondition(state.KcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)
	if kcpCondReady == nil {
		// not yet ready, PV will be created only once KCP NfsInstance is ready
		return nil, nil
	}

	labelsBuilder := util.NewLabelBuilder()
	labelsBuilder.WithCustomLabels(getVolumeLabels(state.ObjAsAwsNfsVolume()))
	labelsBuilder.WithCloudManagerDefaults()

	pvLabels := labelsBuilder.Build()

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   state.Obj().GetNamespace(),
			Name:        getVolumeName(state.ObjAsAwsNfsVolume()),
			Labels:      pvLabels,
			Annotations: getVolumeAnnotations(state.ObjAsAwsNfsVolume()),
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": state.ObjAsAwsNfsVolume().Spec.Capacity,
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   state.ObjAsAwsNfsVolume().Status.Server,
					Path:     "/",
					ReadOnly: false,
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
		},
	}
	err := state.Cluster().K8sClient().Create(ctx, pv)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating PV for AwsNfsVolume", composed.StopWithRequeue, ctx)
	}

	logger.Info("PV for AwsNfsVolume created")

	return nil, nil
}
