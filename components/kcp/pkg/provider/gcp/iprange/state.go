package iprange

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/iprange/types"
	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/iprange/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
)

type State struct {
	types.State

	inSync    bool
	ipAddress string
	prefix    int
	ipRanges  []string

	addressOp    focal.OperationType
	connectionOp focal.OperationType
	curState     v1beta1.StatusState

	address           *compute.Address
	serviceConnection *servicenetworking.Connection

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

func (s State) doesAddressMatch() bool {
	vpc := s.Scope().Spec.Scope.Gcp.VpcNetwork
	return s.address != nil && s.address.Address == s.ipAddress &&
		s.address.PrefixLength == int64(s.prefix) &&
		strings.HasSuffix(s.address.Network, vpc)
}

func (s State) doesConnectionIncludeRange() int {

	if s.serviceConnection == nil {
		return -1
	}

	rangeName := s.ObjAsIpRange().Name
	for i, name := range s.serviceConnection.ReservedPeeringRanges {
		if rangeName == name {
			return i
		}
	}
	return -1
}
