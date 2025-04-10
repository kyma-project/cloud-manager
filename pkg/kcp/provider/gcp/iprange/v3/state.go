package v3

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"

	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	gcpiprangev3client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"

	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type State struct {
	types.State

	computeClient              gcpiprangev3client.ComputeClient
	networkComnnectivityClient gcpiprangev3client.NetworkConnectivityClient

	subnet                  *computepb.Subnetwork
	serviceConnectionPolicy *networkconnectivitypb.ServiceConnectionPolicy
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState types.State) (*State, error)
}

type stateFactory struct {
	computeClientProvider             gcpclient.ClientProvider[gcpiprangev3client.ComputeClient]
	networkConnectivityClientProvider gcpclient.ClientProvider[gcpiprangev3client.NetworkConnectivityClient]
	env                               abstractions.Environment
}

func NewStateFactory(
	computeClientProvider gcpclient.ClientProvider[gcpiprangev3client.ComputeClient],
	networkConnectivityClientProvider gcpclient.ClientProvider[gcpiprangev3client.NetworkConnectivityClient],
	env abstractions.Environment) StateFactory {
	return &stateFactory{
		computeClientProvider:             computeClientProvider,
		networkConnectivityClientProvider: networkConnectivityClientProvider,
		env:                               env,
	}
}

func (statefactory *stateFactory) NewState(ctx context.Context, ipRangeState types.State) (*State, error) {

	computeClient, err := statefactory.computeClientProvider(
		ctx,
		statefactory.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}

	networkConnectivityClient, err := statefactory.networkConnectivityClientProvider(
		ctx,
		statefactory.env.Get("GCP_SA_JSON_KEY_PATH"),
	)
	if err != nil {
		return nil, err
	}

	return newState(ipRangeState, computeClient, networkConnectivityClient), nil
}

func newState(ipRangeState types.State, computeClient gcpiprangev3client.ComputeClient, networkConnectivityClient gcpiprangev3client.NetworkConnectivityClient) *State {
	return &State{
		State:                      ipRangeState,
		computeClient:              computeClient,
		networkComnnectivityClient: networkConnectivityClient,
	}
}
