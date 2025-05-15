package exposedData

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	scopetypes "github.com/kyma-project/cloud-manager/pkg/kcp/scope/types"
)

func NewStateFactory(gcpProvider gcpclient.GcpClientProvider[gcpexposeddataclient.Client]) StateFactory {
	return &stateFactory{
		gcpProvider: gcpProvider,
	}
}

type StateFactory interface {
	NewState(ctx context.Context, baseState scopetypes.State) (composed.State, error)
}

var _ StateFactory = &stateFactory{}

type stateFactory struct {
	gcpProvider gcpclient.GcpClientProvider[gcpexposeddataclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, baseState scopetypes.State) (composed.State, error) {
	gcpClient := f.gcpProvider()

	return &State{
		State:     baseState,
		gcpClient: gcpClient,
	}, nil
}

type State struct {
	scopetypes.State

	gcpClient gcpexposeddataclient.Client

	routers   []*computepb.Router
	addresses []*computepb.Address
}
