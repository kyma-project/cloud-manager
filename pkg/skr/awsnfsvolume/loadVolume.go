package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadVolume(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	vol := &corev1.PersistentVolume{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      state.Obj().GetName(),
	}, vol)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error getting PersistentVolumes", composed.StopWithRequeue, ctx)
	}
	if err != nil {
		state.Volume = vol
	}

	return nil, nil
}
