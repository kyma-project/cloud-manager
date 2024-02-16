package gcpnfsvolume

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
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
	capacity := resource.NewQuantity(int64(nfsVolume.Spec.CapacityGb)*1024*1024*1024, resource.BinarySI)

	//If GcpNfsVolume is not Ready state, continue.
	if !meta.IsStatusConditionTrue(nfsVolume.Status.Conditions, v1beta1.ConditionTypeReady) {
		return nil, nil
	}

	//PV already exists, continue.
	if state.PV != nil {
		return nil, nil
	}

	//If the NFS Host list is empty, create error response.
	if len(state.KcpNfsInstance.Status.Hosts) == 0 {
		logger.WithValues("kyma-name", state.KymaRef).
			WithValues("NfsVolume", state.ObjAsGcpNfsVolume().Name).
			Info("Error creating PV: Not able to get Host(s).")
		return nil, nil
	}

	//Construct a PV Object
	pvName := fmt.Sprintf("%s--%s", nfsVolume.Namespace, nfsVolume.Name)
	state.PV = &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
			Labels: map[string]string{
				v1beta1.LabelNfsVolName: nfsVolume.Name,
				v1beta1.LabelNfsVolNS:   nfsVolume.Namespace,
			},
			Finalizers: []string{
				v1beta1.Finalizer,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			Capacity: v1.ResourceList{
				"storage": *capacity,
			},
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			StorageClassName: "",
			PersistentVolumeSource: v1.PersistentVolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server: state.KcpNfsInstance.Status.Hosts[0],
					Path:   fmt.Sprintf("/%s", nfsVolume.Spec.FileShareName),
				},
			},
		},
	}

	//Create PV
	err := state.SkrCluster.K8sClient().Create(ctx, state.PV)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating PersistentVolume", composed.StopWithRequeue, nil)
	}

	//continue
	return composed.StopWithRequeueDelay(3 * time.Second), nil
}
