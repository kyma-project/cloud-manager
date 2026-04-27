package runtime

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		return nil, nil
	}
}
