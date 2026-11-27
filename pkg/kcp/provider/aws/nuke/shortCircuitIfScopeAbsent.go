package nuke

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// shortCircuitIfScopeAbsent stops the AWS nuke pipeline when the Scope object is not loaded.
// Nuke uses NewWithOptionalScope() so scope may be nil if the Scope was already deleted
// from the cluster. Without it we cannot resolve AWS account/region to build credentials.
func shortCircuitIfScopeAbsent(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.Scope() == nil {
		return composed.StopAndForget, ctx
	}
	return nil, ctx
}
