package iprange

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
	sapiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/iprange/client"
)

type State struct {
	iprangetypes.State
	sapClient sapiprangeclient.Client

	net          *networks.Network
	subnet       *subnets.Subnet
	router       *routers.Router
	routerSubnet *sapclient.RouterSubnetInterfaceInfo
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State) (*State, error)
}

func NewStateFactory(clientProvider sapclient.SapClientProvider[sapiprangeclient.Client]) StateFactory {
	return &stateFactory{
		clientProvider: clientProvider,
	}
}

type stateFactory struct {
	clientProvider sapclient.SapClientProvider[sapiprangeclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, ipRangeState iprangetypes.State) (*State, error) {
	pp := sapclient.NewProviderParamsFromConfig(sapconfig.SapConfig).
		WithDomain(ipRangeState.Scope().Spec.Scope.OpenStack.DomainName).
		WithProject(ipRangeState.Scope().Spec.Scope.OpenStack.TenantName).
		WithRegion(ipRangeState.Scope().Spec.Region)
	sapClient, err := f.clientProvider(ctx, pp)
	if err != nil {
		return nil, fmt.Errorf("error creating sap client for iprange: %w", err)
	}

	return &State{
		State:     ipRangeState,
		sapClient: sapClient,
	}, nil
}

func (s *State) SubnetName() string {
	return fmt.Sprintf("cm-%s", s.ObjAsIpRange().Name)
}
