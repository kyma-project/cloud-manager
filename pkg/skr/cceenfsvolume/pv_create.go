package cceenfsvolume

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
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

	path := state.KcpNfsInstance.Status.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: state.ObjAsCceeNfsVolume().GetPVName(),
			Labels: util.NewLabelBuilder().
				WithCustomLabels(state.ObjAsCceeNfsVolume().GetPVLabels()).
				WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolName, state.ObjAsCceeNfsVolume().Name).
				WithCustomLabel(cloudresourcesv1beta1.LabelNfsVolNS, state.ObjAsCceeNfsVolume().Namespace).
				WithCloudManagerDefaults().
				Build(),
			Annotations: state.ObjAsCceeNfsVolume().GetPVAnnotations(),
			Finalizers: []string{
				api.CommonFinalizerDeletionHook,
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": storageSize,
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   state.KcpNfsInstance.Status.Host,
					Path:     path,
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
	logger.
		WithValues("pvName", pv.Name).
		Info("Created PV for CceeNfsVolume")

	state.PV = pv

	return nil, ctx
}
