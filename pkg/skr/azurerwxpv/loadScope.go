package azurerwxpv

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.Scope() != nil {
		return nil, nil
	}

	scope := &cloudcontrolv1beta1.Scope{}
	err := state.KcpCluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef().Namespace,
		Name:      state.KymaRef().Name,
	}, scope)

	if apierrors.IsNotFound(err) {
		return composed.LogErrorAndReturn(err, "Scope for SKR not found", composed.StopAndForget, ctx)
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP Scope", err, ctx)
	}

	state.SetScope(scope)
	return nil, nil
}
