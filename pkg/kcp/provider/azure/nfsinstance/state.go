package nfsinstance

import (
	"context"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
)

type State struct {
	nfsinstancetypes.State
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error)
}

func NewStateFactory() StateFactory {
	// TODO: implement the state and the factory
	return nil
}
