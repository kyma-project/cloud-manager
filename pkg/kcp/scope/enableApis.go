package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func enableApis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	scope := state.ObjAsScope()
	switch scope.Spec.Provider {
	case cloudcontrolv1beta1.ProviderGCP:
		scope := state.ObjAsScope()
		if len(scope.Status.GcpOperations) == 0 {
			return enableApisGcp(ctx, state)
		} else {
			return checkGcpOperations(ctx, state)
		}
	}

	return nil, nil
}
