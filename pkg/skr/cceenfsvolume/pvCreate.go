package cceenfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func pvCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}

	if state.PV != nil {
		return nil, ctx
	}

	storageSize, err := resource.ParseQuantity(fmt.Sprintf("%dG", state.ObjAsCceeNfsVolume().Spec.CapacityGb))
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error parsing CceeNfsVolume capacity as resource quantity", composed.StopAndForget, ctx)
	}

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: state.ObjAsCceeNfsVolume().GetName(),
			Labels: util.NewLabelBuilder().
				WithCustomLabels(state.ObjAsCceeNfsVolume().GetLabels()).
				WithCloudManagerDefaults().
				Build(),
			Annotations: state.ObjAsCceeNfsVolume().GetAnnotations(),
			Finalizers: []string{
				cloudresourcesv1beta1.Finalizer,
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": storageSize,
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   state.KcpNfsInstance.Status.Host,
					Path:     state.KcpNfsInstance.Status.Path,
					ReadOnly: false,
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
		},
	}

	err = state.Cluster().K8sClient().Create(ctx, pv)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating PV for CceeNfsVolume", composed.StopWithRequeue, ctx)
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Created PV for CceeNfsVolume")

	state.PV = pv

	return nil, ctx
}
