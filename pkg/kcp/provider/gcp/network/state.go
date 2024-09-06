package network

import (
	"context"
	networktypes "github.com/kyma-project/cloud-manager/pkg/kcp/network/types"
)

type State struct {
	networktypes.State
}

type StateFactory interface {
	NewState(ctx context.Context, networkState networktypes.State) (*State, error)
}
