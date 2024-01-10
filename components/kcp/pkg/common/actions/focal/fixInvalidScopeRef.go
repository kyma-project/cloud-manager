package focal

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func fixInvalidScopeRef(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	scopeRef := state.ObjAsCommonObj().ScopeRef()
	scope := state.Scope()
	if scopeRef == nil && scope == nil {
		return nil, nil // whenNoScope will handle this, it will create the Scope, set the reference and requeue
	}
	if scopeRef != nil && scope != nil {
		return nil, nil // all fine, scope from the reference is loaded
	}

	// reference is set, but that scope is not found
	// remove the invalid reference

	err := errors.New("scope that object refers to does not exist")
	logger.WithValues("scopeRef", state.ObjAsCommonObj().ScopeRef().Name).
		Error(err, "Missing scope reference")

	// remove invalid scope reference from the object
	state.ObjAsCommonObj().SetScopeRef(nil)
	err = state.UpdateObj(ctx)
	if err != nil {
		err = fmt.Errorf("error updating object to remove invalid scope reference: %w", err)
		logger.Error(err, "Error updating object")
		return composed.StopWithRequeue, nil
	}

	return composed.StopWithRequeue, nil
}
