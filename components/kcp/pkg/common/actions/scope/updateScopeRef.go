package scope

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func updateScopeRef(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	state.CommonObj().SetScopeRef(&cloudresourcesv1beta1.ScopeRef{
		Name: state.Scope().Name,
	})

	err := state.UpdateObj(ctx)
	if err != nil {
		err = fmt.Errorf("error updating object scope ref: %w", err)
		logger.Error(err, "error saving object with Gcp scope ref")
		return composed.StopWithRequeue, nil // will requeue
	}

	return nil, nil
}
