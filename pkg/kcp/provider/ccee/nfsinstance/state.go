package nfsinstance

import (
	"context"
	"fmt"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	cceeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/client"
	cceeconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/config"
	cceenfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/nfsinstance/client"
)

type State struct {
	nfsinstancetypes.State

	cceeClient cceenfsinstanceclient.Client

	network      *networks.Network
	subnet       *subnets.Subnet
	shareNetwork *sharenetworks.ShareNetwork
	share        *shares.Share
	accessRight  *shares.AccessRight
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error)
}

func NewStateFactory(provider cceeclient.CceeClientProvider[cceenfsinstanceclient.Client]) StateFactory {
	return &stateFactory{
		provider: provider,
	}
}

type stateFactory struct {
	provider cceeclient.CceeClientProvider[cceenfsinstanceclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error) {
	pp := cceeclient.NewProviderParamsFromConfig(cceeconfig.CCEEConfig).
		WithDomain(nfsInstanceState.Scope().Spec.Scope.OpenStack.DomainName).
		WithProject(nfsInstanceState.Scope().Spec.Scope.OpenStack.TenantName).
		WithRegion(nfsInstanceState.Scope().Spec.Region)
	cceeClient, err := f.provider(ctx, pp)
	if err != nil {
		return nil, fmt.Errorf("error creating ccee client: %w", err)
	}

	return &State{
		State:      nfsInstanceState,
		cceeClient: cceeClient,
	}, nil
}

func (s *State) ShareNetworkName() string {
	return fmt.Sprintf("cm-%s", s.Scope().Spec.ShootName)
}

func (s *State) ShareName() string {
	return fmt.Sprintf("cm-%s", s.ObjAsNfsInstance().Name)
}
