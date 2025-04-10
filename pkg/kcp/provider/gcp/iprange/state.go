package iprange

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3"
	gcpiprangev3client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"
)

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (composed.State, error)
}

func NewStateFactory(
	serviceNetworkingClientProvider gcpclient.ClientProvider[gcpiprangeclient.ServiceNetworkingClient],
	computeClientProvider gcpclient.ClientProvider[gcpiprangeclient.ComputeClient],
	v3ComputeClientProvider gcpclient.ClientProvider[gcpiprangev3client.ComputeClient],
	v3NetworkConnectivityClient gcpclient.ClientProvider[gcpiprangev3client.NetworkConnectivityClient],
	env abstractions.Environment,
) StateFactory {
	return &generalStateFactory{
		v2StateFactory: v2.NewStateFactory(serviceNetworkingClientProvider, computeClientProvider, env),
		v3StateFactory: v3.NewStateFactory(v3ComputeClientProvider, v3NetworkConnectivityClient, env),
	}
}

type generalStateFactory struct {
	v2StateFactory v2.StateFactory
	v3StateFactory v3.StateFactory
}

func (f *generalStateFactory) NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (composed.State, error) {
	return nil, errors.New("logical error - not implemented")
}
