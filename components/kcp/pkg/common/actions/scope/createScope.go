package scope

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func createScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	switch state.Provider() {
	case ProviderGCP:
		return createScopeGcp(ctx, state)
	case ProviderAzure:
		return createScopeAzure(ctx, state)
	case ProviderAws:
		return createScopeAws(ctx, state)
	}

	err := fmt.Errorf("unable to handle unknown provider '%s'", state.Provider())
	logger := composed.LoggerFromCtx(ctx)
	logger.Error(err, "Error defining scope")
	return composed.StopAndForget, nil // no requeue

}
