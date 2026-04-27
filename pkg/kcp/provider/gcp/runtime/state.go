package runtime

import (
	"context"

	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

type State struct {
	runtimetypes.State
}

type StateFactory interface {
	NewState(ctx context.Context, runtimeState runtimetypes.State) (*State, error)
}

func NewStateFactory() StateFactory {
	return nil
}
