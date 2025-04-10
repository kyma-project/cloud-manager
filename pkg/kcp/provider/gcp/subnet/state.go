package subnet

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"

	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	gcpClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type State struct {
	focal.State

	computeClient              client.ComputeClient
	networkComnnectivityClient client.NetworkConnectivityClient

	subnet                  *computepb.Subnetwork
	serviceConnectionPolicy *networkconnectivitypb.ServiceConnectionPolicy
}

type StateFactory interface {
	NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
	computeClientProvider             gcpClient.ClientProvider[client.ComputeClient]
	networkConnectivityClientProvider gcpClient.ClientProvider[client.NetworkConnectivityClient]
	env                               abstractions.Environment
}

func NewStateFactory(
	computeClientProvider gcpClient.ClientProvider[client.ComputeClient],
	networkConnectivityClientProvider gcpClient.ClientProvider[client.NetworkConnectivityClient],
	env abstractions.Environment) StateFactory {
	return &stateFactory{
		computeClientProvider:             computeClientProvider,
		networkConnectivityClientProvider: networkConnectivityClientProvider,
		env:                               env,
	}
}

func (statefactory *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {

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

	return newState(focalState, computeClient, networkConnectivityClient), nil
}

func newState(focalState focal.State, computeClient client.ComputeClient, networkConnectivityClient client.NetworkConnectivityClient) *State {
	return &State{
		State:                      focalState,
		computeClient:              computeClient,
		networkComnnectivityClient: networkConnectivityClient,
	}
}

func (s *State) ObjAsGcpSubnet() *cloudcontrolv1beta1.GcpSubnet {
	return s.Obj().(*cloudcontrolv1beta1.GcpSubnet)
}
