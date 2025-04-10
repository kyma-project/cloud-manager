package gcpnfsvolume

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createPersistenceVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	//If NfsVolume is marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR GcpNfsVolume is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}

	//Get GcpNfsVolume object
	nfsVolume := state.ObjAsGcpNfsVolume()
	capacity := gcpNfsVolumeCapacityToResourceQuantity(nfsVolume)

	//If GcpNfsVolume is not Ready state, continue.
	if !meta.IsStatusConditionTrue(nfsVolume.Status.Conditions, v1beta1.ConditionTypeReady) {
		return nil, nil
	}

	//PV already exists, continue.
	if state.PV != nil {
		return nil, nil
	}

	//If the NFS Host list is empty, create error response.
	if len(nfsVolume.Status.Hosts) == 0 {
		logger.WithValues("kyma-name", state.KymaRef).
			WithValues("NfsVolume", state.ObjAsGcpNfsVolume().Name).
			Info("Error creating PV: Not able to get Host(s).")
		return nil, nil
	}

	//Construct a PV Object
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        getVolumeName(nfsVolume),
			Labels:      getVolumeLabels(nfsVolume),
			Annotations: getVolumeAnnotations(nfsVolume),
			Finalizers: []string{
				api.CommonFinalizerDeletionHook,
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": *capacity,
			},
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			StorageClassName: "",
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: nfsVolume.Status.Hosts[0],
					Path:   fmt.Sprintf("/%s", nfsVolume.Spec.FileShareName),
				},
			},
		},
	}

	//Create PV
	err := state.Cluster().K8sClient().Create(ctx, pv)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating PersistentVolume", composed.StopWithRequeue, ctx)
	}

	//continue
	return composed.StopWithRequeueDelay(3 * util.Timing.T1000ms()), nil
}
