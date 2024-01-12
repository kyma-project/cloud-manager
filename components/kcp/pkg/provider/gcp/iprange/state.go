package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/iprange/types"
	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/iprange/client"
	"google.golang.org/api/compute/v1"
)

type State struct {
	types.State

	inSync    bool
	ipAddress string
	prefix    int

	address *compute.Address

	serviceNetworkingClient client.ServiceNetworkingClient
	computeClient           client.ComputeClient
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState types.State) (*State, error)
}

type stateFactory struct {
	serviceNetworkingClientProvider gcpclient.ClientProvider[client.ServiceNetworkingClient]
	computeClientProvider           gcpclient.ClientProvider[client.ComputeClient]
	env                             abstractions.Environment
}

func NewStateFactory(serviceNetworkingClientProvider gcpclient.ClientProvider[client.ServiceNetworkingClient], computeClientProvider gcpclient.ClientProvider[client.ComputeClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		serviceNetworkingClientProvider: serviceNetworkingClientProvider,
		computeClientProvider:           computeClientProvider,
		env:                             env,
	}
}

func (f *stateFactory) NewState(ctx context.Context, ipRangeState types.State) (*State, error) {
	snc, err := f.serviceNetworkingClientProvider(
		ctx,
		f.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}
	cc, err := f.computeClientProvider(
		ctx,
		f.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}

	return newState(ipRangeState, snc, cc), nil
}

func newState(ipRangeState types.State, snc client.ServiceNetworkingClient, cc client.ComputeClient) *State {
	return &State{
		State:                   ipRangeState,
		serviceNetworkingClient: snc,
		computeClient:           cc,
	}
}

func (s State) isMatching(addr *compute.Address) bool {
	vpc := s.Scope().Spec.Scope.Gcp.VpcNetwork
	return addr.Address == s.ipAddress &&
		addr.PrefixLength == int64(s.prefix) &&
		addr.Network == vpc
}
