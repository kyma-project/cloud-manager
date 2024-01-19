package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/nfsinstance/types"
	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/nfsinstance/client"
)

type State struct {
	types.State
	curState        v1beta1.StatusState
	filestoreClient client.FilestoreClient
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState types.State) (*State, error)
}

type stateFactory struct {
	filestoreClientProvider gcpclient.ClientProvider[client.FilestoreClient]
	env                     abstractions.Environment
}

func NewStateFactory(filestoreClientProvider gcpclient.ClientProvider[client.FilestoreClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		filestoreClientProvider: filestoreClientProvider,
		env:                     env,
	}
}

func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState types.State) (*State, error) {
	httpClient, err := gcpclient.GetCachedGcpClient(ctx, f.env.Get("GCP_SA_JSON_KEY_PATH"))
	if err != nil {
		return nil, err
	}
	fc, err := f.filestoreClientProvider(
		ctx,
		httpClient,
	)
	if err != nil {
		return nil, err
	}
	return newState(nfsInstanceState, fc), nil
}

func newState(nfsInstanceState types.State, fc client.FilestoreClient) *State {
	return &State{
		State:           nfsInstanceState,
		filestoreClient: fc,
	}
}
