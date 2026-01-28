package nfsinstance

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
	sapnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/nfsinstance/client"
)

type State struct {
	nfsinstancetypes.State

	sapClient sapnfsinstanceclient.Client

	network      *networks.Network
	subnet       *subnets.Subnet
	shareNetwork *sharenetworks.ShareNetwork
	share        *shares.Share
	accessRight  *sapclient.ShareAccess
}

type StateFactory interface {
	NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error)
}

func NewStateFactory(provider sapclient.SapClientProvider[sapnfsinstanceclient.Client]) StateFactory {
	return &stateFactory{
		provider: provider,
	}
}

type stateFactory struct {
	provider sapclient.SapClientProvider[sapnfsinstanceclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, nfsInstanceState nfsinstancetypes.State) (*State, error) {
	pp := sapclient.NewProviderParamsFromConfig(sapconfig.SapConfig).
		WithDomain(nfsInstanceState.Scope().Spec.Scope.OpenStack.DomainName).
		WithProject(nfsInstanceState.Scope().Spec.Scope.OpenStack.TenantName).
		WithRegion(nfsInstanceState.Scope().Spec.Region)
	sapClient, err := f.provider(ctx, pp)
	if err != nil {
		return nil, fmt.Errorf("error creating sap client for nfs: %w", err)
	}

	return &State{
		State:     nfsInstanceState,
		sapClient: sapClient,
	}, nil
}

func (s *State) ShareNetworkName() string {
	return fmt.Sprintf("cm-%s", s.Scope().Spec.ShootName)
}

func (s *State) ShareName() string {
	return fmt.Sprintf("cm-%s", s.ObjAsNfsInstance().Name)
}
