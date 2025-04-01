package scope

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func scopeCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	switch state.provider {
	case cloudcontrolv1beta1.ProviderGCP:
		return scopeCreateGcp(ctx, state)
	case cloudcontrolv1beta1.ProviderAzure:
		return scopeCreateAzure(ctx, state)
	case cloudcontrolv1beta1.ProviderAws:
		return scopeCreateAws(ctx, state)
	case cloudcontrolv1beta1.ProviderOpenStack:
		return scopeCreateOpenStack(ctx, state)
	}

	err := fmt.Errorf("unable to handle unknown provider '%s'", state.provider)
	logger := composed.LoggerFromCtx(ctx)
	logger.Error(err, "Error defining scope")
	return composed.StopAndForget, nil // no requeue

}
