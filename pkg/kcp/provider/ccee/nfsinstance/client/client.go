package client

import (
	"context"
	"fmt"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	"k8s.io/utils/ptr"
	"net/http"

	cceeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/client"
)

type Client interface {
	ListInternalNetworks(ctx context.Context, name string) ([]networks.Network, error)
	GetNetwork(ctx context.Context, id string) (*networks.Network, error)
	ListSubnets(ctx context.Context, networkId string) ([]subnets.Subnet, error)
	GetSubnet(ctx context.Context, id string) (*subnets.Subnet, error)

	ListShareNetworks(ctx context.Context, networkId string) ([]sharenetworks.ShareNetwork, error)
	GetShareNetwork(ctx context.Context, id string) (*sharenetworks.ShareNetwork, error)
	CreateShareNetwork(ctx context.Context, networkId, subnetId, name string) (*sharenetworks.ShareNetwork, error)
	DeleteShareNetwork(ctx context.Context, id string) error

	ListShares(ctx context.Context, shareNetworkId string) ([]shares.Share, error)
	GetShare(ctx context.Context, id string) (*shares.Share, error)
	CreateShare(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*shares.Share, error)
	DeleteShare(ctx context.Context, id string) error
	ListShareExportLocations(ctx context.Context, id string) ([]shares.ExportLocation, error)

	ListShareAccessRights(ctx context.Context, id string) ([]shares.AccessRight, error)
	GrantShareAccess(ctx context.Context, shareId string, cidr string) (*shares.AccessRight, error)
	RevokeShareAccess(ctx context.Context, shareId, accessId string) error
}

var _ Client = &client{}

type client struct {
	svc *gophercloud.ServiceClient
}

func NewClientProvider() cceeclient.CceeClientProvider[Client] {
	return func(ctx context.Context, pp cceeclient.ProviderParams) (Client, error) {
		svc, err := cceeclient.NewProviderClient(ctx, pp)
		if err != nil {
			return nil, fmt.Errorf("failed to create new ccee provider client: %v", err)
		}
		return &client{svc: svc}, nil
	}
}

func (c *client) ListInternalNetworks(ctx context.Context, name string) ([]networks.Network, error) {
	pg, err := networks.List(c.svc, external.ListOptsExt{
		External: ptr.To(false),
		ListOptsBuilder: networks.ListOpts{
			Name: name,
		},
	}).AllPages(ctx)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing private networks: %w", err)
	}
	arr, err := networks.ExtractNetworks(pg)
	if err != nil {
		return nil, fmt.Errorf("error extracting private networks: %w", err)
	}
	return arr, nil
}

func (c *client) GetNetwork(ctx context.Context, id string) (*networks.Network, error) {
	n, err := networks.Get(ctx, c.svc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	return n, err
}

func (c *client) ListSubnets(ctx context.Context, networkId string) ([]subnets.Subnet, error) {
	pg, err := subnets.List(c.svc, subnets.ListOpts{
		NetworkID: networkId,
	}).AllPages(ctx)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing subnets: %w", err)
	}
	arr, err := subnets.ExtractSubnets(pg)
	if err != nil {
		return nil, fmt.Errorf("error extracting subnets: %w", err)
	}

	return arr, nil
}

func (c *client) GetSubnet(ctx context.Context, id string) (*subnets.Subnet, error) {
	subnet, err := subnets.Get(ctx, c.svc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting subnet: %w", err)
	}

	return subnet, nil
}

// Share Networks ------------------------------------------------------------------------------

func (c *client) ListShareNetworks(ctx context.Context, networkId string) ([]sharenetworks.ShareNetwork, error) {
	pg, err := sharenetworks.ListDetail(c.svc, sharenetworks.ListOpts{
		NeutronNetID: networkId,
	}).AllPages(ctx)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing sharenetworks: %w", err)
	}
	arr, err := sharenetworks.ExtractShareNetworks(pg)
	if err != nil {
		return nil, fmt.Errorf("error extracting sharenetworks: %w", err)
	}

	return arr, nil
}

func (c *client) GetShareNetwork(ctx context.Context, id string) (*sharenetworks.ShareNetwork, error) {
	net, err := sharenetworks.Get(ctx, c.svc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting sharenetwork: %w", err)
	}
	return net, nil
}

func (c *client) CreateShareNetwork(ctx context.Context, networkId, subnetId, name string) (*sharenetworks.ShareNetwork, error) {
	net, err := sharenetworks.Create(ctx, c.svc, sharenetworks.CreateOpts{
		NeutronNetID:    networkId,
		NeutronSubnetID: subnetId,
		Name:            name,
	}).Extract()
	if err != nil {
		return net, fmt.Errorf("error creating sharenetwork: %w", err)
	}
	return net, nil
}

func (c *client) DeleteShareNetwork(ctx context.Context, id string) error {
	err := sharenetworks.Delete(ctx, c.svc, id).ExtractErr()
	if err != nil {
		return fmt.Errorf("error deleting sharenetwork: %w", err)
	}
	return nil
}

// shares ----------------------------------------------------------------------------------

// share.status possible values https://docs.openstack.org/manila/latest/user/create-and-manage-shares.html
// These “-ing” states end in a “available” state if everything goes well. They may end up in an “error” state in case there is an issue.
// * available
// * error
// * creating
// * extending
// * shrinking
// * migrating

func (c *client) ListShares(ctx context.Context, shareNetworkId string) ([]shares.Share, error) {
	pg, err := shares.ListDetail(c.svc, shares.ListOpts{
		ShareNetworkID: shareNetworkId,
	}).AllPages(ctx)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error listing shares: %w", err)
	}
	arr, err := shares.ExtractShares(pg)
	if err != nil {
		return nil, fmt.Errorf("error extracting shares: %w", err)
	}
	return arr, nil
}

func (c *client) GetShare(ctx context.Context, id string) (*shares.Share, error) {
	sh, err := shares.Get(ctx, c.svc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting share: %w", err)
	}
	return sh, nil
}

func (c *client) CreateShare(ctx context.Context, shareNetworkId, name string, size int, snapshotID string, metadata map[string]string) (*shares.Share, error) {
	sh, err := shares.Create(ctx, c.svc, shares.CreateOpts{
		ShareProto:     "NFS",
		Size:           size,
		Name:           name,
		ShareNetworkID: shareNetworkId,
		SnapshotID:     snapshotID,
		Metadata:       metadata,
	}).Extract()
	if err != nil {
		return sh, fmt.Errorf("error creating share: %w", err)
	}
	return sh, nil
}

func (c *client) DeleteShare(ctx context.Context, id string) error {
	err := shares.Delete(ctx, c.svc, id).ExtractErr()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting share: %w", err)
	}
	return nil
}

func (c *client) ListShareExportLocations(ctx context.Context, id string) ([]shares.ExportLocation, error) {
	arr, err := shares.ListExportLocations(ctx, c.svc, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("error listing export locations: %w", err)
	}
	return arr, nil
}

// share access -------------------------------------------------------------------

func (c *client) ListShareAccessRights(ctx context.Context, id string) ([]shares.AccessRight, error) {
	arr, err := shares.ListAccessRights(ctx, c.svc, id).Extract()
	if err != nil {
		return nil, fmt.Errorf("error listing access rights: %w", err)
	}
	return arr, nil
}

func (c *client) GrantShareAccess(ctx context.Context, shareId string, cidr string) (*shares.AccessRight, error) {
	ar, err := shares.GrantAccess(ctx, c.svc, shareId, shares.GrantAccessOpts{
		AccessType:  "ip",
		AccessTo:    cidr,
		AccessLevel: "rw",
	}).Extract()
	if err != nil {
		return ar, fmt.Errorf("error granting access to share: %w", err)
	}
	return ar, nil
}

func (c *client) RevokeShareAccess(ctx context.Context, shareId, accessId string) error {
	err := shares.RevokeAccess(ctx, c.svc, shareId, shares.RevokeAccessOpts{
		AccessID: accessId,
	}).ExtractErr()
	if err != nil {
		return fmt.Errorf("error revoking access to share: %w", err)
	}
	return nil
}
