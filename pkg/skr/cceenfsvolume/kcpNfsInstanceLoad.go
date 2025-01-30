package cceenfsvolume

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func kcpNfsInstanceLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.ObjAsCceeNfsVolume().Status.Id == "" {
		return composed.LogErrorAndReturn(errors.New("Missing CceeNfsVolume status.id"), "Logical error", composed.StopAndForget, ctx)
	}

	kcpNfsInstnace := &cloudcontrolv1beta1.NfsInstance{}

	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsCceeNfsVolume().Status.Id,
	}, kcpNfsInstnace)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP NfsInstance for CceeNfsVolume", composed.StopWithRequeue, ctx)
	}
	if err != nil {
		state.KcpNfsInstance = kcpNfsInstnace
	}

	return nil, ctx
}
