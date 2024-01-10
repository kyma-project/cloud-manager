package nfsinstance

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-resources-manager/components/lib/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		panic(errors.New("not implemented"))
	}
}
