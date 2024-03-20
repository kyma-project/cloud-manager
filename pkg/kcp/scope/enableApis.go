package scope

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func enableApis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	scope := state.ObjAsScope()
	switch scope.Spec.Provider {
	case cloudcontrolv1beta1.ProviderGCP:
		scope := state.ObjAsScope()
		if scope.Status.GcpOperations == nil || len(scope.Status.GcpOperations) == 0 {
			return enableApisGcp(ctx, state)
		} else {
			return checkGcpOperations(ctx, state)
		}
	case cloudcontrolv1beta1.ProviderAzure:
		return nil, nil
	case cloudcontrolv1beta1.ProviderAws:
		return nil, nil
	}

	err := fmt.Errorf("unable to handle unknown provider '%s'", state.provider)
	logger := composed.LoggerFromCtx(ctx)
	logger.Error(err, "Error in enableApis")
	// It is not this action's responsibility to handle unknown providers. This is already handled in the createScope action.
	return nil, nil // no requeue
}
