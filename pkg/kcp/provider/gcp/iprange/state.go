package iprange

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	v2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
	v3 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3"
	v3client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"
)

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (composed.State, error)
}

func NewStateFactory(
	serviceNetworkingClientProvider gcpclient.ClientProvider[iprangeclient.ServiceNetworkingClient],
	computeClientProvider gcpclient.ClientProvider[iprangeclient.ComputeClient],
	v3ComputeClientProvider gcpclient.ClientProvider[v3client.ComputeClient],
	v3NetworkConnectivityClient gcpclient.ClientProvider[v3client.NetworkConnectivityClient],
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
