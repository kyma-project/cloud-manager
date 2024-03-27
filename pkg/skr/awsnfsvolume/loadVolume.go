package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	vol := &corev1.PersistentVolume{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      getVolumeName(state.ObjAsAwsNfsVolume()),
	}, vol)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error getting PersistentVolume by getVolumeName()", composed.StopWithRequeue, ctx)
	}

	if apierrors.IsNotFound(err) {
		// first PVs were created with name = AwsNfsVolume.status.id
		// next, a feature was added in AwsNfsVolume to specify PV name
		// this is a fallback to old behavior where PV.name = AwsNfsVolume.status.id
		// to remain compatibility with already created PVs
		err = state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
			Namespace: state.Obj().GetNamespace(),
			Name:      state.ObjAsAwsNfsVolume().Status.Id,
		}, vol)
		if client.IgnoreNotFound(err) != nil {
			return composed.LogErrorAndReturn(err, "Error getting PersistentVolume by status.id", composed.StopWithRequeue, ctx)
		}
	}

	if err == nil {
		state.Volume = vol
	}

	return nil, nil
}
