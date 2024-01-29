package nfsinstance

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		panic(errors.New("not implemented"))
	}
}
