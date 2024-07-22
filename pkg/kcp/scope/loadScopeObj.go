package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func loadScopeObj(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	err := state.LoadObj(ctx)
	if apierrors.IsNotFound(err) {
		// continue to create one
		return nil, nil
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Scope object", composed.StopWithRequeue, ctx)
	}
	return nil, nil
}
