package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/iprange/types"
	"strings"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	gcpclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/iprange/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
)

type State struct {
	types.State

	inSync        bool
	ipAddress     string
	prefix        int
	ipRanges      []string
	projectNumber int64

	addressOp    gcpclient.OperationType
	connectionOp gcpclient.OperationType
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
	httpClient, err := gcpclient.GetCachedGcpClient(ctx, f.env.Get("GCP_SA_JSON_KEY_PATH"))
	if err != nil {
		return nil, err
	}
	snc, err := f.serviceNetworkingClientProvider(
		ctx,
		httpClient,
	)
	if err != nil {
		return nil, err
	}
	cc, err := f.computeClientProvider(
		ctx,
		httpClient,
	)
	if err != nil {
		return nil, err
	}
	projectId := ipRangeState.Scope().Spec.Scope.Gcp.Project
	projectNumber, err := gcpclient.GetCachedProjectNumber(ctx, projectId, httpClient)
	if err != nil {
		return nil, err
	}
	return newState(ipRangeState, snc, cc, projectNumber), nil
}

func newState(ipRangeState types.State, snc client.ServiceNetworkingClient, cc client.ComputeClient, projectNumber int64) *State {
	return &State{
		State:                   ipRangeState,
		serviceNetworkingClient: snc,
		computeClient:           cc,
		projectNumber:           projectNumber,
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
