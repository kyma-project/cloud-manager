package v2

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
	"strings"
)

type State struct {
	iprangetypes.State

	inSync    bool
	ipAddress string
	prefix    int
	ipRanges  []string

	addressOp    gcpclient.OperationType
	connectionOp gcpclient.OperationType
	curState     v1beta1.StatusState

	address           *compute.Address
	serviceConnection *servicenetworking.Connection

	serviceNetworkingClient gcpiprangeclient.ServiceNetworkingClient
	computeClient           gcpiprangeclient.ComputeClient
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State) (*State, error)
}

type stateFactory struct {
	serviceNetworkingClientProvider gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	computeClientProvider           gcpclient.ClientProvider[gcpiprangeclient.ComputeClient]
	env                             abstractions.Environment
}

func NewStateFactory(serviceNetworkingClientProvider gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient], computeClientProvider gcpclient.ClientProvider[gcpiprangeclient.ComputeClient], env abstractions.Environment) StateFactory {
	return &stateFactory{
		serviceNetworkingClientProvider: serviceNetworkingClientProvider,
		computeClientProvider:           computeClientProvider,
		env:                             env,
	}
}

func (f *stateFactory) NewState(ctx context.Context, ipRangeState iprangetypes.State) (*State, error) {

	snc, err := f.serviceNetworkingClientProvider(
		ctx,
		gcpclient.GcpConfig.CredentialsFile,
	)
	if err != nil {
		return nil, err
	}
	cc, err := f.computeClientProvider(
		ctx,
		gcpclient.GcpConfig.CredentialsFile,
	)
	if err != nil {
		return nil, err
	}

	return newState(ipRangeState, snc, cc), nil
}

func newState(ipRangeState iprangetypes.State, snc gcpiprangeclient.ServiceNetworkingClient, cc gcpiprangeclient.ComputeClient) *State {
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
	if s.address == nil {
		return -1
	}

	for i, name := range s.serviceConnection.ReservedPeeringRanges {
		if s.address.Name == name {
			return i
		}
	}
	return -1
}
