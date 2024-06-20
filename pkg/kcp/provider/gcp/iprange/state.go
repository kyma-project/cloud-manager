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
	v1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v1"
	v2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v2"
)

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (composed.State, error)
}

func NewStateFactory(serviceNetworkingClientProvider gcpclient.ClientProvider[iprangeclient.ServiceNetworkingClient], computeClientProvider gcpclient.ClientProvider[iprangeclient.ComputeClient], env abstractions.Environment) StateFactory {
	return &generalStateFactory{
		v1StateFactory: v1.NewStateFactory(serviceNetworkingClientProvider, computeClientProvider, env),
		v2StateFactory: v2.NewStateFactory(serviceNetworkingClientProvider, computeClientProvider, env),
	}
}

type generalStateFactory struct {
	v1StateFactory v1.StateFactory
	v2StateFactory v2.StateFactory
}

func (f *generalStateFactory) NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (composed.State, error) {
	return nil, errors.New("logical error - not implemented")
	//return f.v1StateFactory.NewState(ctx, ipRangeState, logger)
}
