package nfsinstance

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(errors.New("azure/nfsinstance should not been called"), "Logical error")
		return nil, nil
	}
}
