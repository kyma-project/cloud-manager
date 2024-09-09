package v1

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	client2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	client3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
	"strings"
)

type State struct {
	types.State

	inSync    bool
	ipAddress string
	prefix    int
	ipRanges  []string

	addressOp    client2.OperationType
	connectionOp client2.OperationType
	curState     v1beta1.StatusState

	address           *compute.Address
	serviceConnection *servicenetworking.Connection

	serviceNetworkingClient client3.ServiceNetworkingClient
	computeClient           client3.ComputeClient
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState types.State) (*State, error)
}

type stateFactory struct {
	serviceNetworkingClientProvider client2.ClientProvider[client3.ServiceNetworkingClient]
	computeClientProvider           client2.ClientProvider[client3.ComputeClient]
	env                             abstractions.Environment
}

func NewStateFactory(serviceNetworkingClientProvider client2.ClientProvider[client3.ServiceNetworkingClient], computeClientProvider client2.ClientProvider[client3.ComputeClient], env abstractions.Environment) StateFactory {
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

func newState(ipRangeState types.State, snc client3.ServiceNetworkingClient, cc client3.ComputeClient) *State {
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

	rangeName := s.ObjAsIpRange().Spec.RemoteRef.Name
	for i, name := range s.serviceConnection.ReservedPeeringRanges {
		if rangeName == name {
			return i
		}
	}
	return -1
}
