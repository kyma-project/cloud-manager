package composed

import (
	"context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func LoadObj(ctx context.Context, state State) (error, context.Context) {
	err := state.LoadObj(ctx)
	if apierrors.IsNotFound(err) {
		return StopAndForget, nil
	}
	if err != nil {
		return LogErrorAndReturn(err, "Error loading object", StopWithRequeue, ctx)
	}

	return nil, nil
}

func LoadObjNoStopIfNotFound(ctx context.Context, state State) (error, context.Context) {
	err := state.LoadObj(ctx)
	if apierrors.IsNotFound(err) {
		return nil, ctx
	}
	if err != nil {
		return LogErrorAndReturn(err, "Error loading object", StopWithRequeue, ctx)
	}

	return nil, nil
}

func IsObjLoaded(ctx context.Context, state State) bool {
	if state.Obj() == nil || state.Obj().GetName() == "" || state.Obj().GetResourceVersion() == "" {
		return false
	}
	return true
}
